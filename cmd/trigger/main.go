package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Tinee/hackathon2018/asdasd"
	"github.com/Tinee/hackathon2018/repository"
	"github.com/globalsign/mgo"
	"github.com/tinee/hackathon2018/domain"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Request from IFTTT
type Request struct {
	Triggers struct {
		Token string `json:"trigger_identity"`
		From  string `json:"hours_start"`
		To    string `json:"hours_stop"`
		Limit int    `json:"limit"`
	} `json:"triggerFields"`
}

type Result struct {
	Data Data `json:"data"`
}

type Data struct {
	From string          `json:"from"`
	To   string          `json:"to"`
	Mix  Generationmixes `json:"generationmix"`
}

type Generationmix struct {
	Fuel       string  `json:"fuel"`
	Percentage float32 `json:"perc"`
}

type Generationmixes []Generationmix

// AggregateGreenEnergy calculates green energy percentage
func (g Generationmixes) AggregateGreenEnergy() (res float32) {
	for _, element := range g {
		switch element.Fuel {
		case "solar", "hydro", "wind":
			res += element.Percentage
		}
	}
	return res
}

// WithHourMinute sets the hour and minute on a given Time
func WithHourMinute(now time.Time, hmString string) (time.Time, error) {
	hourStr := hmString[:2]
	minuteStr := hmString[3:]

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return time.Time{}, err
	}
	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, now.Second(), now.Nanosecond(), now.Location()), nil
}

// BuildResponse builds response to IFTTT
func BuildResponse(isOverLimit bool, greenPercentage float32) ([]byte, error) {

	return json.Marshal(domain.Response{
		Data: []domain.ResponseDetail{
			domain.ResponseDetail{
				IsOverLimit:     isOverLimit,
				GreenPercentage: greenPercentage,
				CreatedAt:       time.Now().UTC().Format(time.RFC3339),
				Meta: domain.Meta{
					Id:        "14b9-1fd2-acaa-5df5",
					Timestamp: int(time.Now().Unix()),
				},
			},
		},
	})
}

func ConnectToDatabase(dbAddr string) (repository.EventRepository, error) {
	if dbAddr == "" {
		fmt.Println("No DB_ADDR provided")
		return nil, errors.New("No DB_ADDR provided")
	}

	fmt.Println("Connecting to DB")

	dialInfo, err := mgo.ParseURL(dbAddr)
	dialInfo.Timeout = 30 * time.Second
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}

	mongoClient, err := repository.NewMongoClient(dialInfo)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println("Initialising Events Repo")
	eventRepo, err := repository.NewMongoEventsRespository(mongoClient, "events")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return eventRepo, nil
}

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Starting the application...")

	errr := auth.ValidateIFTTTRequest(e)
	if errr != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	now := time.Now()
	var req Request
	err := json.Unmarshal([]byte(e.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	tos := req.Triggers.To
	froms := req.Triggers.From
	limit := req.Triggers.Limit
	token := req.Triggers.Token

	from, err := WithHourMinute(now, froms)
	if err != nil {
		fmt.Printf("Could not parse from  %s %s\n", froms, err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	to, err := WithHourMinute(now, tos)
	if err != nil {
		fmt.Printf("Could not parse to  %s %s\n", tos, err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	if now.Before(from) || now.After(to) {
		fmt.Println("Exiting early outside of range")
		body, _ := BuildResponse(false, 0.1)

		return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
			"content-type": "application/json; charset=utf-8",
		}}, nil
	}

	repo, err := ConnectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		fmt.Printf("DB connection failure %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	existingEvents, err := repo.FindAllByTokenIdentity(token, limit)
	if err != nil {
		fmt.Printf("Error when FindAllByTokenIdentity %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	fmt.Println(existingEvents)

	response, err := http.Get("http://api.carbonintensity.org.uk/generation")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	jsonData, _ := ioutil.ReadAll(response.Body)

	var result Result
	err = json.Unmarshal(jsonData, &result)

	if err != nil {
		fmt.Printf("Error Unmarshaling json %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	aggregation := result.Data.Mix.AggregateGreenEnergy()

	isHigher := aggregation > 30.0

	body, err := BuildResponse(isHigher, aggregation)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
		"content-type": "application/json; charset=utf-8",
	}}, nil
}

func main() {
	lambda.Start(Handle)
}

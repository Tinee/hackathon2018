package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Tinee/hackathon2018/asdasd"
	"github.com/Tinee/hackathon2018/domain"
	"github.com/Tinee/hackathon2018/repository"
	"github.com/Tinee/hackathon2018/service"
	"github.com/globalsign/mgo"

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

// Temp empty response
func EmptyResponse() ([]byte, error) {

	return json.Marshal(domain.Response{
		Data: []domain.ResponseDetail{
			domain.ResponseDetail{
				IsOverLimit:     false,
				GreenPercentage: 0.0,
				CreatedAt:       time.Now().UTC().Format(time.RFC3339),
				Meta: domain.Meta{
					Id:        "14b9-1fd2-acaa-5df5",
					Timestamp: time.Now().Unix(),
				},
			},
		},
	})
}

func BuildResponse(events_ *[]domain.Event) ([]byte, error) {
	events := *events_

	details := make([]domain.ResponseDetail, len(events))

	for i, event := range events {
		details[i] = event.AsResponseDetail()
	}

	return json.Marshal(domain.Response{
		Data: details,
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

	var req Request
	err := json.Unmarshal([]byte(e.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	to := req.Triggers.To
	from := req.Triggers.From
	limit := req.Triggers.Limit
	token := req.Triggers.Token

	// Handle trigger Window
	inTriggerWindow, err := service.InTriggerWindow(from, to)
	if err != nil {
		fmt.Printf("Error when determining if inside error window %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	if !inTriggerWindow {
		fmt.Println("Exiting early outside of range")
		body, _ := EmptyResponse()

		return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
			"content-type": "application/json; charset=utf-8",
		}}, nil
	}
	//

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

	body, err := BuildResponse(existingEvents)
	if err != nil {
		fmt.Printf("Failed to build response %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	// QUICK
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
		"content-type": "application/json; charset=utf-8",
	}}, nil

	// response, err := http.Get("http://api.carbonintensity.org.uk/generation")
	// if err != nil {
	// 	fmt.Printf("The HTTP request failed with error %s\n", err)
	// 	return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	// }

	// jsonData, _ := ioutil.ReadAll(response.Body)

	// var result Result
	// err = json.Unmarshal(jsonData, &result)

	// if err != nil {
	// 	fmt.Printf("Error Unmarshaling json %s\n", err)
	// 	return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	// }

	// aggregation := result.Data.Mix.AggregateGreenEnergy()

	// isHigher := aggregation > 30.0

	// body, err := BuildResponse(isHigher, aggregation)
	// if err != nil {
	// 	fmt.Printf("The HTTP request failed with error %s\n", err)
	// 	return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	// }

	// return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
	// 	"content-type": "application/json; charset=utf-8",
	// }}, nil
}

func main() {
	lambda.Start(Handle)
}

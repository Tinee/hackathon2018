package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Tinee/hackathon2018/asdasd"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Request from IFTTT
type Request struct {
	triggers struct {
		from string `json:"hours_start"`
		to   string `json:"hours_stop"`
	} `json:"triggerFields"`
}

// Response to IFTTT
type Response struct {
	data []ResponseDetail `json:"data"`
}

type ResponseDetail struct {
	isOverLimit     bool    `json:"isOverLimit"`
	greenPercentage float32 `json:"greenPercentage"`
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
	return json.Marshal(Response{
		data: []ResponseDetail{
			ResponseDetail{
				isOverLimit:     isOverLimit,
				greenPercentage: greenPercentage,
			},
		},
	})
}

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Starting the application...")
	// if errResp := auth.ValidateIFTTTRequest(e); errResp != nil {
	// 	return *errResp, nil
	// }

	errr := auth.ValidateIFTTTRequest(e)
	if errr != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	now := time.Now()
	var req Request
	err := json.Unmarshal([]byte(e.Body), &req)

	tos := req.triggers.to
	froms := req.triggers.from
	fmt.Println(froms)
	fmt.Println(tos)

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
		body, _ := BuildResponse(false, 0.0)

		return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
	}

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

	// aggregation := result.Data.Mix.AggregateGreenEnergy()

	// isHigher := aggregation > 30.0

	// body, err := BuildResponse(isHigher, aggregation)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: "hejsan", StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handle)
}

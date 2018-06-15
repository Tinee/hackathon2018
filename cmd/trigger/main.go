package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

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

func (g Generationmixes) AggregateGreenEnergy() (res float32) {
	for _, element := range g {
		switch element.Fuel {
		case "solar", "hydro", "wind":
			res += element.Percentage
		}
	}
	return res
}

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

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Starting the application...")

	now := time.Now()
	fmt.Println(now)

	// To take from args
	fromString := "00:30"
	toString := "14:30"

	from, err := WithHourMinute(now, fromString)
	if err != nil {
		fmt.Printf("Could not parse from  %s %s\n", fromString, err)
		// Exit early here
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	to, err := WithHourMinute(now, toString)
	if err != nil {
		fmt.Printf("Could not parse to  %s %s\n", toString, err)
		// Exit early here
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	if now.Before(from) || now.After(to) {
		fmt.Println("Exiting early outside of range")
		return events.APIGatewayProxyResponse{Body: "false", StatusCode: 200}, nil
	}

	response, err := http.Get("http://api.carbonintensity.org.uk/generation")

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		// Exit early here
		return events.APIGatewayProxyResponse{Body: "false", StatusCode: 200}, nil
	}

	jsonData, _ := ioutil.ReadAll(response.Body)

	var result Result
	err = json.Unmarshal(jsonData, &result)

	if err != nil {
		fmt.Printf("Error %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	aggregation := result.Data.Mix.AggregateGreenEnergy()

	isHigher := strconv.FormatBool(aggregation > 30.0)

	return events.APIGatewayProxyResponse{Body: isHigher, StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handle)
}

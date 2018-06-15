package main

import (
	"encoding/json"
	"fmt"

	"github.com/Tinee/hackathon2018/asdasd"
	"github.com/Tinee/hackathon2018/domain"
	"github.com/Tinee/hackathon2018/service"

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

// Temp empty response
func EmptyResponse() ([]byte, error) {

	return json.Marshal(domain.Response{
		Data: []domain.ResponseDetail{},
	})
}

func BuildResponse(events []domain.ResponseDetail) ([]byte, error) {
	return json.Marshal(domain.Response{
		Data: events,
	})
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

	// If there are events in the DB then return those
	existingEvents, err := service.ExistingEvents(token, limit)
	if err != nil {
		fmt.Printf("Error getting the existing events %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	fmt.Println(existingEvents)

	if len(existingEvents) != 0 {
		body, err := BuildResponse(existingEvents)
		if err != nil {
			fmt.Printf("Failed to build response %s\n", err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, nil
		}

		return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
			"content-type": "application/json; charset=utf-8",
		}}, nil
	}

	// If no existing events and outside window then return an empty array
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

	// Otherwise lookup the generation
	aggregation, err := service.LookupGreenEnergyPercentage()
	if err != nil {
		fmt.Printf("Error when Looking up green energy percentage %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	isHigher := aggregation > 30.0

	responseDetail, err := service.SaveNewEvent(token, isHigher, aggregation)
	if err != nil {
		fmt.Printf("Error Saving the new event %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	// TODO nicer??
	details := make([]domain.ResponseDetail, 1)
	details[0] = responseDetail

	body, err := BuildResponse(details)
	if err != nil {
		fmt.Printf("Failed to build response %s\n", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
		"content-type": "application/json; charset=utf-8",
	}}, nil
}

func main() {
	lambda.Start(Handle)
}

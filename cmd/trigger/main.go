package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Tinee/hackathon2018/asdasd"
	"github.com/Tinee/hackathon2018/domain"
	"github.com/Tinee/hackathon2018/service"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Request from IFTTT
type Request struct {
	TriggerIdentity string `json:"trigger_identity"`
	Triggers        struct {
		From string `json:"hours_start"`
		To   string `json:"hours_stop"`
	} `json:"triggerFields"`
	Limit int `json:"limit"`
}

func EmptyResponse() ([]byte, error) {

	return json.Marshal(domain.Response{
		Data: []domain.ResponseDetail{},
	})
}

func ErrorResponse(err error) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       "{\"errors\": [{\"message\": \"" + err.Error() + "\"}]}",
		StatusCode: 400,
		Headers: map[string]string{
			"content-type": "application/json; charset=utf-8",
		},
	}
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

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Starting the application...")
	fmt.Printf("Body %s ", e.Body)

	errr := auth.ValidateIFTTTRequest(e)
	if errr != nil {
		return *errr, nil
	}

	req := Request{}
	req.Limit = -1 // TODO

	err := json.Unmarshal([]byte(e.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	fmt.Println("Validating request")
	to := req.Triggers.To
	if to == "" {
		err = errors.New("Missing to")
		return ErrorResponse(err), nil
	}
	from := req.Triggers.From
	if from == "" {
		err = errors.New("Missing from")
		return ErrorResponse(err), nil
	}
	triggerIdentity := req.TriggerIdentity
	if triggerIdentity == "" {
		err = errors.New("Missing triggerIdentity")
		return ErrorResponse(err), nil
	}

	limit := req.Limit

	if limit == 0 {
		fmt.Println("Limit is actually 0 exiting early")
		body, _ := EmptyResponse()

		return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
			"content-type": "application/json; charset=utf-8",
		}}, nil
	}

	if limit == -1 {
		limit = 0
	}

	results_, err := service.HandleEvent(from, to, triggerIdentity, limit)
	if err != nil {
		return ErrorResponse(err), nil
	}
	body, err := BuildResponse(results_)
	if err != nil {
		return ErrorResponse(err), nil
	}
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200, Headers: map[string]string{
		"content-type": "application/json; charset=utf-8",
	}}, nil
}

func main() {
	lambda.Start(Handle)
}

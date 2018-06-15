package main

import (
	"encoding/json"
	"log"

	auth "github.com/Tinee/hackathon2018/asdasd"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type (
	TestSetupResponse struct {
		Data struct {
			Samples struct {
				Triggers struct {
					LumoHours struct {
						HoursStart string `json:"hours_start"`
						HoursStop  string `json:"hours_stop"`
					} `json:"lumo_hours"`
				} `json:"triggers"`
			} `json:"samples"`
		} `json:"data"`
	}
)

func BuildTestResponse(hoursStart, hoursStop string) ([]byte, error) {
	resp := TestSetupResponse{}
	resp.Data.Samples.Triggers.LumoHours.HoursStart = hoursStart
	resp.Data.Samples.Triggers.LumoHours.HoursStop = hoursStop
	log.Print(resp)
	return json.Marshal(resp)
}

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	errResp := auth.ValidateIFTTTRequest(e)
	if errResp != nil {
		return *errResp, nil
	}

	resp, _ := BuildTestResponse("04:00", "23:30")
	output := string(resp)
	log.Print(output)
	return events.APIGatewayProxyResponse{Body: output, StatusCode: 200, Headers: map[string]string{
		"content-type": "application/json; charset=utf-8",
	}}, nil
}

func main() {
	lambda.Start(Handle)
}

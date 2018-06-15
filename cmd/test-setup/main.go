package main

import (
	"encoding/json"

	auth "github.com/Tinee/hackathon2018/auth"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type (
	TestSetupResponse struct {
		data struct {
			samples struct {
				triggers struct {
					lumoHours struct {
						hoursStart string `json:"hour_start"`
						hoursStop  string `json:"hour_stop"`
					} `json:"lumo_hours"`
				} `json:"triggers"`
			} `json:"samples"`
		} `json:"data"`
	}
)

func BuildTestResponse(hoursStart, hoursStop string) ([]byte, error) {
	resp := TestSetupResponse{}
	resp.data.samples.triggers.lumoHours.hoursStart = hoursStart
	resp.data.samples.triggers.lumoHours.hoursStop = hoursStop
	return json.Marshal(resp)
}

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	errResp := auth.ValidateIFTTTRequest(e)
	if errResp != nil {
		return *errResp, nil
	}

	resp, _ := BuildTestResponse("22:00", "23:30")

	return events.APIGatewayProxyResponse{Body: string(resp), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handle)
}

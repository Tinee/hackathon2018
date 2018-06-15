package main

import (
	"encoding/json"

	"github.com/Tinee/hackathon2018/asdasd"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type (
	TriggerOptionsResponse struct {
		Data []TriggerOption `json:"data"`
	}

	TriggerOption struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}
)

var options = []TriggerOption{
	TriggerOption{Label: "00:00", Value: "00:00"},
	TriggerOption{Label: "00:30", Value: "00:30"},
	TriggerOption{Label: "01:00", Value: "01:00"},
	TriggerOption{Label: "01:30", Value: "01:30"},
	TriggerOption{Label: "02:00", Value: "02:00"},
	TriggerOption{Label: "02:30", Value: "02:30"},
	TriggerOption{Label: "03:00", Value: "03:00"},
	TriggerOption{Label: "03:30", Value: "03:30"},
	TriggerOption{Label: "04:00", Value: "04:00"},
	TriggerOption{Label: "04:30", Value: "04:30"},
	TriggerOption{Label: "05:00", Value: "05:00"},
	TriggerOption{Label: "05:30", Value: "05:30"},
	TriggerOption{Label: "06:00", Value: "06:00"},
	TriggerOption{Label: "06:30", Value: "06:30"},
	TriggerOption{Label: "07:00", Value: "07:00"},
	TriggerOption{Label: "07:30", Value: "07:30"},
	TriggerOption{Label: "08:00", Value: "08:00"},
	TriggerOption{Label: "08:30", Value: "08:30"},
	TriggerOption{Label: "09:00", Value: "09:00"},
	TriggerOption{Label: "09:30", Value: "09:30"},
	TriggerOption{Label: "10:00", Value: "10:00"},
	TriggerOption{Label: "10:30", Value: "10:30"},
	TriggerOption{Label: "11:00", Value: "11:00"},
	TriggerOption{Label: "11:30", Value: "11:30"},
	TriggerOption{Label: "12:00", Value: "12:00"},
	TriggerOption{Label: "12:30", Value: "12:30"},
	TriggerOption{Label: "13:00", Value: "13:00"},
	TriggerOption{Label: "13:30", Value: "13:30"},
	TriggerOption{Label: "14:00", Value: "14:00"},
	TriggerOption{Label: "14:30", Value: "14:30"},
	TriggerOption{Label: "15:00", Value: "15:00"},
	TriggerOption{Label: "15:30", Value: "15:30"},
	TriggerOption{Label: "16:00", Value: "16:00"},
	TriggerOption{Label: "16:30", Value: "16:30"},
	TriggerOption{Label: "17:00", Value: "17:00"},
	TriggerOption{Label: "17:30", Value: "17:30"},
	TriggerOption{Label: "18:00", Value: "18:00"},
	TriggerOption{Label: "18:30", Value: "18:30"},
	TriggerOption{Label: "19:00", Value: "19:00"},
	TriggerOption{Label: "19:30", Value: "19:30"},
	TriggerOption{Label: "20:00", Value: "20:00"},
	TriggerOption{Label: "20:30", Value: "20:30"},
	TriggerOption{Label: "21:00", Value: "21:00"},
	TriggerOption{Label: "21:30", Value: "21:30"},
	TriggerOption{Label: "22:00", Value: "22:00"},
	TriggerOption{Label: "22:30", Value: "22:30"},
	TriggerOption{Label: "23:00", Value: "23:00"},
	TriggerOption{Label: "23:30", Value: "23:30"},
}

var b, err = json.Marshal(TriggerOptionsResponse{Data: options})
var response = string(b)

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	errResp := auth.ValidateIFTTTRequest(e)
	if errResp != nil {
		return *errResp, nil
	}

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "\"" + err.Error() + "\""}, nil
	}
	return events.APIGatewayProxyResponse{Body: response, StatusCode: 200, Headers: map[string]string{
		"content-type": "application/json; charset=utf-8",
	}}, nil
}

func main() {
	lambda.Start(Handle)
}

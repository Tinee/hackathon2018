package main

import (
	auth "github.com/Tinee/hackathon2018/asdasd"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	errResp := auth.ValidateIFTTTRequest(e)
	if errResp != nil {
		return *errResp, nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handle)
}

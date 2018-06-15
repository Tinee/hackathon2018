package main

import (
	"log"
	"os"

	auth "github.com/Tinee/hackathon2018/auth"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println(e.Headers["ifttt-service-key"])
	log.Println(os.Getenv("IFTTT_SERVICE_KEY"))
	err := auth.ValidateServiceKey(e.Headers["ifttt-service-key"])
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "\"" + err.Error() + "\"", StatusCode: 401}, nil
	}
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handle)
}

package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var result []string
	for i := 0; i < 24; i++ {
		result = append(result, fmt.Sprintf("%d:00", i), fmt.Sprintf("%d:30", i))
	}

	return events.APIGatewayProxyResponse{Body: "Hejsan", StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handle)
}

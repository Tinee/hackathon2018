package auth

import (
	"errors"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

func ValidateIFTTTRequest(e events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	log.Print(e.Headers)
	err := ValidateServiceKey(e.Headers["Ifttt-Service-Key"])
	if err != nil {
		return &events.APIGatewayProxyResponse{Body: "\"" + err.Error() + "\"", StatusCode: 401}
	}
	return nil
}

func ValidateServiceKey(key string) error {
	if os.Getenv("IFTTT_SERVICE_KEY") == "" {
		return errors.New("No environment variable specified when deploying this lambda")
	}
	if key == "" {
		return errors.New("No Service Key Given, Instead Given")
	}
	if key != os.Getenv("IFTTT_SERVICE_KEY") {
		return errors.New("Invalid Service Key")
	}
	return nil
}

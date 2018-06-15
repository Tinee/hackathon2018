package auth

import (
	"errors"
	"os"
)

func ValidateServiceKey(key string) error {
	if os.Getenv("IFTTT_SERVICE_KEY") == "" {
		return errors.New("No environment variable specified when deploying this lambda")
	}
	if key == "" {
		return errors.New("No Service Key Given")
	}
	if key != os.Getenv("IFTTT_SERVICE_KEY") {
		return errors.New("Invalid Service Key")
	}
	return nil
}

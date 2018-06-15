package auth

import (
	"errors"
	"os"
)

func ValidateServiceKey(key string) error {
	if key != os.Getenv("IFTTT_KEY") {
		return errors.New("Invalid Service Key")
	}
	return nil
}

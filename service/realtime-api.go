package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type (
	RealtimeIdRequest struct {
		Data []RealtimeId `json:"data"`
	}
	RealtimeId struct {
		UserID    string `json:"userId,omitempty"`
		TriggerID string `json:"trigger_identity,omitempty"`
	}
)

func UpdateIFTTT(triggerIdentifiers []string) error {
	var realtimeIds []RealtimeId
	for _, elem := range triggerIdentifiers {
		realtimeIds = append(realtimeIds, RealtimeId{TriggerID: elem})
	}

	jsonB, err := json.Marshal(RealtimeIdRequest{Data: realtimeIds})

	log.Println("Sending to IFTTT")
	log.Println(string(jsonB))

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://realtime.ifttt.com/v1/notifications", bytes.NewBuffer(jsonB))

	if err != nil {
		return err
	}

	req.Header.Set("IFTTT-Service-Key", os.Getenv("IFTTT_SERVICE_KEY"))
	req.Header.Set("X-Request-ID", RandStringRunes(53))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	log.Println("Response from IFTTT")
	log.Println(string(body))

	if err != nil {
		fmt.Printf("The HTTP couldn't be read %s\n", err)
		return err
	}

	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Tinee/hackathon2018/repository"
	"github.com/Tinee/hackathon2018/service"
	"github.com/globalsign/mgo"

	"github.com/Tinee/hackathon2018/domain"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type NewMockData struct {
	MockData    domain.MockData
	IdsToUpdate []string
}

func ConnectToDatabase(dbAddr string) (repository.MockDataRepository, repository.EventRepository, error) {
	if dbAddr == "" {
		fmt.Println("No DB_ADDR provided")
		return nil, nil, errors.New("No DB_ADDR provided")
	}

	fmt.Println("Connecting to DB")

	dialInfo, err := mgo.ParseURL(dbAddr)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}
	dialInfo.Timeout = 30 * time.Second

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}

	mongoClient, err := repository.NewMongoClient(dialInfo)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	fmt.Println("Initialising MockData Repo")
	mockDataRepo, err := repository.NewMongoMockDataRespository(mongoClient, "mockData")
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	fmt.Println("Initialising Events Repo")
	eventsRepo, err := repository.NewMongoEventsRespository(mongoClient, "events")
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	return mockDataRepo, eventsRepo, nil
}

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var req domain.MockData
	err := json.Unmarshal([]byte(e.Body), &req)
	fmt.Println("Unmarshalled")
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "\"" + err.Error() + "\""}, nil
	}

	mockDataRepo, eventsRepo, err := ConnectToDatabase(os.Getenv("DB_ADDR"))

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "\"" + err.Error() + "\""}, nil
	}
	fmt.Println("Connected to DB")

	mockData, err := mockDataRepo.Set(req)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "\"" + err.Error() + "\""}, nil
	}

	fmt.Println("Set Mock Data")

	ids, err := eventsRepo.FindUnique(1000)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "\"" + err.Error() + "\""}, nil
	}

	err = service.UpdateIFTTT(*ids)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "\"" + err.Error() + "\""}, nil
	}

	b, err := json.Marshal(NewMockData{MockData: *mockData, IdsToUpdate: *ids})

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(b), Headers: map[string]string{
		"Access-Control-Allow-Origin": "*",
	}}, nil
}

func main() {
	lambda.Start(Handle)
}

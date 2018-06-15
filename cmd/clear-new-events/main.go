package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Tinee/hackathon2018/repository"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/globalsign/mgo"
)

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

	_, eventsRepo, err := ConnectToDatabase(os.Getenv("DB_ADDR"))
	eventsRepo.ClearAllNewEvents()
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handle)
}

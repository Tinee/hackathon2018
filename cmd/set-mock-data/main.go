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
	"github.com/globalsign/mgo"

	"github.com/Tinee/hackathon2018/domain"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func ConnectToDatabase(dbAddr string) (repository.MockDataRepository, error) {
	if dbAddr == "" {
		fmt.Println("No DB_ADDR provided")
		return nil, errors.New("No DB_ADDR provided")
	}

	fmt.Println("Connecting to DB")

	dialInfo, err := mgo.ParseURL(dbAddr)
	dialInfo.Timeout = 30 * time.Second
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}

	mongoClient, err := repository.NewMongoClient(dialInfo)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println("Initialising MockData Repo")
	mockDataRepo, err := repository.NewMongoMockDataRespository(mongoClient, "mockData")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return mockDataRepo, nil
}

func Handle(e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var req domain.MockData
	err := json.Unmarshal([]byte(e.Body), &req)
	fmt.Println("Unmarshalled")
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "\"" + err.Error() + "\""}, nil
	}

	mockDataRepo, err := ConnectToDatabase(os.Getenv("DB_ADDR"))

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "\"" + err.Error() + "\""}, nil
	}
	fmt.Println("Connected to DB")

	mockData, err := mockDataRepo.Set(req)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "\"" + err.Error() + "\""}, nil
	}
	fmt.Println("Set Mock Data")

	b, err := json.Marshal(mockData)

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(b)}, nil
}

func main() {
	lambda.Start(Handle)
}

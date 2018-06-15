package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Tinee/hackathon2018/domain"
	"github.com/Tinee/hackathon2018/repository"
	"github.com/globalsign/mgo"
)

// from as 12:00 to as 12:30
func InTriggerWindow(from string, to string) (bool, error) {
	now := time.Now()

	fromParsed, err := withHourMinute(now, from)
	if err != nil {
		fmt.Printf("Could not parse from  %s %s\n", from, err)
		return false, err
	}

	toParsed, err := withHourMinute(now, to)
	if err != nil {
		fmt.Printf("Could not parse to  %s %s\n", to, err)
		return false, err
	}

	return (now.After(fromParsed) && now.Before(toParsed)), nil
}

func ExistingEventsForToday(token string, limit int) ([]domain.ResponseDetail, error) {
	repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		fmt.Printf("DB connection failure %s\n", err)
		return nil, err
	}

	events_, err := repo.FindAllByTokenSinceBeginningOfDay(token, time.Now(), limit)
	if err != nil {
		fmt.Printf("Error when FindAllByTokenIdentity %s\n", err)
		return nil, err
	}

	events := *events_
	fmt.Println(events)

	details := make([]domain.ResponseDetail, len(events))
	for i, event := range events {
		details[i] = event.AsResponseDetail()
	}

	return details, nil
}

func LookupGreenEnergyPercentage() (float32, error) {
	response, err := http.Get("http://api.carbonintensity.org.uk/generation")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return -1, err
	}

	jsonData, _ := ioutil.ReadAll(response.Body)

	var result domain.GenerationResult
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		fmt.Printf("Error Unmarshaling json %s\n", err)
		return -1, err
	}

	return result.Data.Mix.AggregateGreenEnergy(), nil
}

func SaveNewEvent(triggerIdentity string, isOverLimit bool, greenPercentage float32) (domain.ResponseDetail, error) {
	event := domain.Event{
		TriggerIdentity: triggerIdentity,
		IsOverLimit:     isOverLimit,
		GreenPercentage: greenPercentage,
	}

	hasFirstEvent, err := HasFirstEvent(triggerIdentity)
	if err != nil {
		fmt.Printf("HasFirstEvent failure %s\n", err)
		return domain.ResponseDetail{}, err
	}

	if hasFirstEvent {
		event.CreatedAt = time.Now()
	}

	repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		fmt.Printf("DB connection failure %s\n", err)
		return domain.ResponseDetail{}, err
	}

	saved, err := repo.Insert(event)
	if err != nil {
		fmt.Printf("Error Inserting into DB %s\n", err)
		return domain.ResponseDetail{}, err
	}

	return saved.AsResponseDetail(), nil
}

func HasFirstEvent(triggerIdentity string) (bool, error) {
	repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		fmt.Printf("DB connection failure %s\n", err)
		return false, err
	}

	events_, err := repo.FindAllByTokenIdentity(triggerIdentity, 1)
	if err != nil {
		fmt.Printf("FindAllByTokenIdentity failure %s\n", err)
		return false, err
	}

	events := *events_
	return (len(events) > 0), nil
}

// Privates

func connectToDatabase(dbAddr string) (repository.EventRepository, error) {
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

	fmt.Println("Initialising Events Repo")
	eventRepo, err := repository.NewMongoEventsRespository(mongoClient, "events")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return eventRepo, nil
}

// WithHourMinute sets the hour and minute on a given Time
func withHourMinute(now time.Time, hmString string) (time.Time, error) {
	hourStr := hmString[:2]
	minuteStr := hmString[3:]

	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return time.Time{}, err
	}
	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, now.Second(), now.Nanosecond(), now.Location()), nil
}

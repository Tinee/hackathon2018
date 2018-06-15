package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Tinee/hackathon2018/domain"
	"github.com/Tinee/hackathon2018/repository"
	"github.com/globalsign/mgo"
)

func HandleEvent(from string, to string, triggerIdentity string, limit int) (*[]domain.Event, error) {
	fmt.Println("Handling event")

	fmt.Println("EnsureInitialExists")
	err := EnsureInitialExists(triggerIdentity)
	if err != nil {
		return nil, err
	}

	fmt.Println("AlreadyExistsForToday")
	alreadyExistsForToday, err := AlreadyExistsForToday(triggerIdentity)
	if err != nil {
		return nil, err
	}
	if alreadyExistsForToday {
		return FindAll(triggerIdentity, limit)
	}

	fmt.Println("IsWithinTimeWindow")
	withinTimeRange, err := IsWithinTimeWindow(from, to)
	if err != nil {
		return nil, err
	}
	if !withinTimeRange {
		return FindAll(triggerIdentity, limit)
	}

	// time.Time{} needs to be replaced
	fmt.Println("LookupGreenEnergyPercentage")
	greenPercetageNow, err := LookupGreenEnergyPercentage(time.Time{})
	if greenPercetageNow > 30.0 {
		fmt.Println("SaveNewEvent")
		SaveNewEvent(triggerIdentity, greenPercetageNow)
	}

	fmt.Println("FindAll")
	return FindAll(triggerIdentity, limit)
}

func EnsureInitialExists(triggerIdentity string) error {
	repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		return err
	}

	events_, err := repo.FindAllByTokenIdentity(triggerIdentity, 1)
	if err != nil {
		return err
	}
	events := *events_

	if len(events) == 0 {
		fmt.Println("Initial event missing ")

		event := domain.Event{
			TriggerIdentity: triggerIdentity,
			IsOverLimit:     false,
			GreenPercentage: 0.0,
		}

		_, err := repo.Insert(event)
		if err != nil {
			return err
		}
	}

	return nil
}

func AlreadyExistsForToday(triggerIdentity string) (bool, error) {
	repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		return false, err
	}

	events_, err := repo.FindAllByTokenSinceBeginningOfDay(triggerIdentity, time.Now(), 1)
	if err != nil {
		return false, err
	}
	events := *events_

	return len(events) > 0, nil
}

func FindAll(tokenIdentity string, limit int) (*[]domain.Event, error) {
	repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		return nil, err
	}

	return repo.FindAllByTokenIdentity(tokenIdentity, limit)
}

func IsWithinTimeWindow(from string, to string) (bool, error) {
	now := time.Now()

	fromParsed, err := withHourMinute(now, from)
	if err != nil {
		return false, err
	}

	toParsed, err := withHourMinute(now, to)
	if err != nil {
		return false, err
	}

	return (now.After(fromParsed) && now.Before(toParsed)), nil
}

func LookupGreenEnergyPercentage(now time.Time) (float32, error) {
	if (now == time.Time{}) {
		fmt.Println("lookupNormalGreenEnergyPercentage")
		return lookupNormalGreenEnergyPercentage()
	}
	fmt.Println("LookupGreenEnergyPercentage")
	return lookupMockGreenEnergyPercentage(now)
}

func lookupNormalGreenEnergyPercentage() (float32, error) {
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

func lookupMockGreenEnergyPercentage(now time.Time) (float32, error) {
	log.Println("Getting mock data")
	nowFormatted := now.Format("2006-01-02T15:04Z")
	toFormatted := now.Add(time.Hour).Format("2006-01-02T15:04Z")
	log.Println("Now time: " + nowFormatted)
	log.Println("To time: " + toFormatted)
	url := "http://api.carbonintensity.org.uk/generation/" + nowFormatted + "/" + toFormatted
	log.Println("url: " + url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		return -1, err
	}

	jsonData, _ := ioutil.ReadAll(response.Body)

	var result domain.GenerationSeriesResult
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		fmt.Printf("Error Unmarshaling json %s\n", err)
		return -1, err
	}
	for _, elem := range result.Data {
		log.Println(elem.From)
		fromTime, err := time.Parse("2006-01-02T15:04Z", elem.From)
		if err != nil {
			return 0, err
		}
		nowHrs := now.Hour()
		nowMins := now.Minute()
		fromHrs := fromTime.Hour()
		fromMins := fromTime.Minute()
		if nowHrs == fromHrs && nowMins == fromMins {
			log.Println("Found hours")
			log.Println(nowHrs)
			log.Println(fromHrs)
			log.Println("Found mins")
			log.Println(nowMins)
			log.Println(fromMins)
			return elem.Mix.AggregateGreenEnergy(), nil
		}
	}
	return 0, errors.New("could not find a valid time range in the mock data")
}

func SaveNewEvent(triggerIdentity string, greenPercentage float32) error {
	event := domain.Event{
		TriggerIdentity: triggerIdentity,
		IsOverLimit:     true, // As we only call this when saving a normal trigger
		GreenPercentage: greenPercentage,
		CreatedAt:       time.Now(),
	}

	repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		return err
	}

	_, err = repo.Insert(event)
	return err
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

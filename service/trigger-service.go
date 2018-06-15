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

	fmt.Println("getMockedDate")
	mockTime, err := getMockedDate()
	if err != nil {
		return nil, err
	}

	fmt.Println("AlreadyExistsForToday")
	alreadyExistsForToday, err := AlreadyExistsForToday(triggerIdentity, mockTime)
	if err != nil {
		return nil, err
	}
	if alreadyExistsForToday {
		return FindAll(triggerIdentity, limit)
	}

	fmt.Println("IsWithinTimeWindow")
	withinTimeRange, err := IsWithinTimeWindow(from, to, mockTime)
	if err != nil {
		return nil, err
	}
	if !withinTimeRange {
		return FindAll(triggerIdentity, limit)
	}

	fmt.Println("LookupGreenEnergyPercentage")
	greenPercetageNow, err := LookupGreenEnergyPercentage(mockTime)
	if err != nil {
		return nil, err
	}

	fmt.Println("isNearEndOfWindow")
	nearEnd, err := isNearEndOfWindow(to, mockTime)
	if err != nil {
		return nil, err
	}

	if greenPercetageNow >= 50.0 || nearEnd {
		fmt.Printf("SaveNewEvent greenPercent [%s] nearEnd [%s]", greenPercetageNow, nearEnd)
		SaveNewEvent(triggerIdentity, greenPercetageNow, mockTime)
	}

	fmt.Println("FindAll")
	return FindAll(triggerIdentity, limit)
}

func EnsureInitialExists(triggerIdentity string) error {
	_, repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
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

func AlreadyExistsForToday(triggerIdentity string, now time.Time) (bool, error) {
	if (now == time.Time{}) {
		now = time.Now()
	}

	_, repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		return false, err
	}

	events_, err := repo.FindAllByTokenSinceBeginningOfDay(triggerIdentity, now, 1)
	if err != nil {
		return false, err
	}
	events := *events_

	return len(events) > 0, nil
}

func getMockedDate() (time.Time, error) {
	mockDataRepo, _, err := connectToDatabase(os.Getenv("DB_ADDR"))
	mockData, err := mockDataRepo.Get()
	if err == mgo.ErrNotFound {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return mockData.Now, nil
}

func FindAll(tokenIdentity string, limit int) (*[]domain.Event, error) {
	_, repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		return nil, err
	}

	return repo.FindAllByTokenIdentity(tokenIdentity, limit)
}

func IsWithinTimeWindow(from string, to string, now time.Time) (bool, error) {

	if (now == time.Time{}) {
		now = time.Now()
	}

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

func isNearEndOfWindow(to string, now time.Time) (bool, error) {

	if (now == time.Time{}) {
		now = time.Now()
	}

	thirtyMins, err := time.ParseDuration("30m")
	if err != nil {
		return false, err
	}

	nowPlus30Mins := now.Add(thirtyMins)

	toParsed, err := withHourMinute(now, to)
	if err != nil {
		return false, err
	}

	return nowPlus30Mins.After(toParsed), nil
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

func SaveNewEvent(triggerIdentity string, greenPercentage float32, now time.Time) error {
	if (now == time.Time{}) {
		now = time.Now()
	}
	event := domain.Event{
		TriggerIdentity: triggerIdentity,
		IsOverLimit:     true, // As we only call this when saving a normal trigger
		GreenPercentage: greenPercentage,
		CreatedAt:       now,
	}

	_, repo, err := connectToDatabase(os.Getenv("DB_ADDR"))
	if err != nil {
		return err
	}

	_, err = repo.Insert(event)
	return err
}

// Privates

func connectToDatabase(dbAddr string) (repository.MockDataRepository, repository.EventRepository, error) {
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

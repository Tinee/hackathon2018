package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Tinee/hackathon2018/domain"
)

func main() {
	lookupTime, e := time.Parse("2006-01-02T15:04Z", "2018-06-13T11:30Z")
	f, e := LookupMockGreenEnergyPercentage(lookupTime)
	if e != nil {
		panic(e)
	}
	log.Println("found")
	log.Println(f)
}

func LookupMockGreenEnergyPercentage(now time.Time) (float32, error) {
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

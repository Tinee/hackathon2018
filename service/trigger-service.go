package service

import (
	"fmt"
	"strconv"
	"time"
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

package domain

import "time"

type Event struct {
	ID              string
	triggerIdentity string
	isOverLimit     bool
	greenPercentage float32
	CreatedAt       time.Time
}

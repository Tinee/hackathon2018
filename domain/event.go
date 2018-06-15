package domain

import "time"

type Event struct {
	ID              string
	triggerIdentity string
	isOverLimit     bool
	greenPercentage float32
	CreatedAt       time.Time
}

// Response to IFTTT
type Response struct {
	Data []ResponseDetail `json:"data"`
}

type ResponseDetail struct {
	IsOverLimit     bool    `json:"isOverLimit"`
	GreenPercentage float32 `json:"greenPercentage"`
	CreatedAt       string  `json:"created_at"`
	Meta            Meta    `json:"meta"`
}

type Meta struct {
	Id        string `json:"id"`
	Timestamp int    `json:"timestamp"`
}

package domain

import "time"

type Event struct {
	ID              string `bson:"_id"`
	TriggerIdentity string
	IsOverLimit     bool
	GreenPercentage float32
	CreatedAt       time.Time
}

func (e Event) asResponseDetail() ResponseDetail {
	return ResponseDetail{
		IsOverLimit:     e.IsOverLimit,
		CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05.999999+01:00"),
		GreenPercentage: e.GreenPercentage,
		Meta: Meta{
			Id:        e.ID,
			Timestamp: e.CreatedAt.Unix(),
		},
	}
}

type MockData struct {
	ID              string `bson:"_id"`
	Now             time.Time
	StartDataPeriod time.Time
	EndDataPeriod   time.Time
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
	Timestamp int64  `json:"timestamp"`
}

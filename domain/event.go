package domain

import "time"

type Event struct {
	ID              string    `bson:"_id"`
	TriggerIdentity string    `bson:"triggerIdentity"`
	IsOverLimit     bool      `bson:"isOverLimit"`
	GreenPercentage float32   `bson:"greenPercentage"`
	CreatedAt       time.Time `bson:"createdAt"`
}

func (e Event) AsResponseDetail() ResponseDetail {
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
	ID              string    `bson:"_id"`
	Now             time.Time `bson:"now"`
	StartDataPeriod time.Time `bson:"startDataPeriod"`
	EndDataPeriod   time.Time `bson:"endDataPeriod"`
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

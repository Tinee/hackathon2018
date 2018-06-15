package repository

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Event struct {
	ID              string
	triggerIdentity string
	isOverLimit     bool
	greenPercentage float32
	CreatedAt       time.Time
}

type EventRepository interface {
	Insert(event Event) (*Event, error)
	FindAllByTokenIdentity(token string, limit int) (*[]Event, error)
}

type mongoEventRepository struct {
	client     *mongoClient
	collection string
}

func NewMongoEventsRespository(client *mongoClient, collection string) (EventRepository, error) {
	repo := &mongoEventRepository{
		client:     client,
		collection: collection,
	}

	s := client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	err := coll.EnsureIndex(mgo.Index{
		Key:        []string{"triggerIdentity"},
		Background: false,
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *mongoEventRepository) Insert(event Event) (*Event, error) {

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	event.ID = bson.NewObjectId().String()
	event.CreatedAt = time.Now()
	err := coll.Insert(event)

	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (repo *mongoEventRepository) FindAllByTokenIdentity(tokenIdentity string, limit int) (*[]Event, error) {

	if limit == 0 {
		limit = 49
	}

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	var results []Event
	err := coll.Find(bson.M{
		"tokenIdentity": tokenIdentity,
	}).All(&results)

	if err == mgo.ErrNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &results, nil
}

package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/Tinee/hackathon2018/domain"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type EventRepository interface {
	Insert(event domain.Event) (*domain.Event, error)
	FindUnique(limit int) (*[]domain.Event, error)
	FindAllByTokenIdentity(token string, limit int) (*[]domain.Event, error)
	FindAllByTokenIdentityBefore(tokenIdentity string, endPeriod time.Time, limit int) (*[]domain.Event, error)
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

func (repo *mongoEventRepository) Insert(event domain.Event) (*domain.Event, error) {

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

func (repo *mongoEventRepository) FindUnique(limit int) (*[]domain.Event, error) {

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	log.Println("Getting find unique")
	var results []domain.Event
	err := coll.Find(bson.M{}).All(&results)

	if err == mgo.ErrNotFound {
		fmt.Println("Not Found")
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	log.Print(results)
	return &results, nil
}

func (repo *mongoEventRepository) FindAllByTokenIdentityBefore(
	tokenIdentity string,
	endPeriod time.Time,
	limit int,
) (*[]domain.Event, error) {

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	var results []domain.Event
	err := coll.Find(bson.M{
		"triggerIdentity": tokenIdentity,
		"createdAt":       bson.M{"$lt": endPeriod},
	}).Limit(limit).All(&results)

	if err == mgo.ErrNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &results, nil
}

func (repo *mongoEventRepository) FindAllByTokenIdentity(tokenIdentity string, limit int) (*[]domain.Event, error) {

	if limit == 0 {
		limit = 50
	}

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	var results []domain.Event
	err := coll.Find(bson.M{
		"triggerIdentity": tokenIdentity,
	}).Limit(limit).All(&results)

	if err == mgo.ErrNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &results, nil
}

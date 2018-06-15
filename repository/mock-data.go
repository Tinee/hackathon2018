package repository

import (
	"log"

	"github.com/Tinee/hackathon2018/domain"
	"github.com/globalsign/mgo/bson"
)

type MockDataRepository interface {
	Set(mockData domain.MockData) (*domain.MockData, error)
	Get() (*domain.MockData, error)
	Clear() error
}

type mongoMockDataRepository struct {
	client     *mongoClient
	collection string
}

func NewMongoMockDataRespository(client *mongoClient, collection string) (MockDataRepository, error) {
	repo := &mongoMockDataRepository{
		client:     client,
		collection: collection,
	}

	s := client.session.Copy()
	defer s.Close()

	return repo, nil
}

func (repo *mongoMockDataRepository) Set(mockData domain.MockData) (*domain.MockData, error) {

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	coll.DropCollection()
	log.Print("Dropped collection")
	mockData.ID = bson.NewObjectId().String()
	err := coll.Insert(mockData)

	if err != nil {
		return nil, err
	}
	log.Print("Inserted")
	return &mockData, nil
}

func (repo *mongoMockDataRepository) Get() (*domain.MockData, error) {

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	var doc domain.MockData
	err := coll.Find(bson.M{}).One(&doc)

	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (repo *mongoMockDataRepository) Clear() error {

	s := repo.client.session.Copy()
	defer s.Close()
	coll := s.DB("").C(repo.collection)

	return coll.DropCollection()
}

package repository

import (
	"github.com/Tinee/hackathon2018/domain"
	"github.com/globalsign/mgo/bson"
)

type MockDataRepository interface {
	Set(mockData domain.MockData) (*domain.MockData, error)
	Get() (*domain.MockData, error)
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
	mockData.ID = bson.NewObjectId().String()
	err := coll.Insert(mockData)

	if err != nil {
		return nil, err
	}
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

package main

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStorage struct {
	client *mongo.Client
	ctx    context.Context
}

const dbName = "test"
const collectionName = "persons"
const mongoUri = "mongodb://root:password@localhost:27017"

func NewMongoStorage() (*MongoStorage, error) {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, err
	}

	return &MongoStorage{client, ctx}, nil
}

func (s *MongoStorage) Close() {
	s.client.Disconnect(s.ctx)
}

func (s *MongoStorage) GetAll() ([]*Person, error) {
	collection := s.client.Database(dbName).Collection(collectionName)

	cursor, _ := collection.Find(s.ctx, bson.D{})

	mps := []*MongoPerson{}
	err := cursor.All(s.ctx, &mps)
	if err != nil {
		return nil, err
	} else if len(mps) == 0 {
		return nil, personNotFoundError
	}
	var pp []*Person
	for _, mp := range mps {
		pp = append(pp, mp.toPerson())
	}

	return pp, nil
}

func (s *MongoStorage) Add(p *Person) (*Person, error) {
	_, err := s.GetPersonByID(p.ID)
	if err == nil {
		return nil, personExistError
	}

	collection := s.client.Database(dbName).Collection(collectionName)
	mp := p.toMongoPerson()
	res, err := collection.InsertOne(s.ctx, mp)
	if err != nil {
		return nil, err
	}
	res = res
	p, err = s.GetPersonByID(p.ID)
	return p, nil
}

func (s *MongoStorage) GetPersonByID(id uuid.UUID) (*Person, error) {
	collection := s.client.Database(dbName).Collection(collectionName)

	cursor, _ := collection.Find(s.ctx, bson.D{{"_id", id.String()}})

	var mps []*MongoPerson
	err := cursor.All(s.ctx, &mps)
	if err != nil {
		return nil, err
	} else if len(mps) == 0 {
		return nil, personNotFoundError
	}

	return mps[0].toPerson(), nil
}

func (s *MongoStorage) GetPersonsByName(name string) ([]*Person, error) {
	collection := s.client.Database(dbName).Collection(collectionName)

	cursor, _ := collection.Find(s.ctx, bson.D{{"name", name}})

	var mps []*MongoPerson
	err := cursor.All(s.ctx, &mps)
	if err != nil {
		return nil, err
	} else if len(mps) == 0 {
		return nil, personNotFoundError
	}

	var pp []*Person
	for _, mp := range mps {
		pp = append(pp, mp.toPerson())
	}
	return pp, nil
}

func (s *MongoStorage) GetPersonsByCommunication(value string) ([]*Person, error) {
	collection := s.client.Database(dbName).Collection(collectionName)
	cursor, _ := collection.Find(s.ctx, bson.D{{"communication.value", value}})

	mps := []*MongoPerson{}
	err := cursor.All(s.ctx, &mps)
	if err != nil {
		return nil, err
	} else if len(mps) == 0 {
		return nil, personNotFoundError
	}
	var pp []*Person
	for _, mp := range mps {
		pp = append(pp, mp.toPerson())
	}

	return pp, nil
}

func (s *MongoStorage) UpdatePerson(person *Person) (*Person, error) {
	_, err := s.GetPersonByID(person.ID)
	if err != nil {
		return nil, err
	}

	collection := s.client.Database(dbName).Collection(collectionName)
	mp := person.toMongoPerson()
	_, err = collection.ReplaceOne(s.ctx, bson.D{{"_id", mp.ID}}, mp)
	if err != nil {
		return nil, err
	}
	p, err := s.GetPersonByID(person.ID)
	return p, err
}

func (s *MongoStorage) DeletePerson(id uuid.UUID) (*Person, error) {
	p, err := s.GetPersonByID(id)
	if err != nil {
		return nil, err
	}

	collection := s.client.Database(dbName).Collection(collectionName)
	_, err = collection.DeleteOne(s.ctx, bson.D{{"_id", id.String()}})
	if err != nil {
		return nil, err
	}
	p, err = s.GetPersonByID(id)
	return p, err
}

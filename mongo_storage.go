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

func (s *MongoStorage) GetAll() []*Person {
	collection := s.client.Database(dbName).Collection(collectionName)

	cursor, _ := collection.Find(s.ctx, bson.D{})

	var mps []*MongoPerson
	err := cursor.All(s.ctx, &mps)
	if err != nil || len(mps) == 0 {
		return nil
	}
	var p []*Person
	for _, mp := range mps {
		p = append(p, mp.toPerson())
	}

	return p
}

func (s *MongoStorage) Add(p *Person) error {
	if s.GetPersonById(p.ID) != nil {
		return personExistError
	}

	collection := s.client.Database(dbName).Collection(collectionName)
	mp := p.toMongoPerson()
	_, err := collection.InsertOne(s.ctx, mp)
	return err
}

func (s *MongoStorage) GetPersonById(id uuid.UUID) *Person {
	collection := s.client.Database(dbName).Collection(collectionName)

	cursor, _ := collection.Find(s.ctx, bson.D{{"_id", id.String()}})

	var mps []*MongoPerson
	err := cursor.All(s.ctx, &mps)
	if err != nil || len(mps) == 0 {
		return nil
	}

	return mps[0].toPerson()
}

func (s *MongoStorage) GetPersonsByName(name string) []*Person {
	collection := s.client.Database(dbName).Collection(collectionName)

	cursor, _ := collection.Find(s.ctx, bson.D{{"name", name}})

	var mps []*MongoPerson
	err := cursor.All(s.ctx, &mps)
	if err != nil || len(mps) == 0 {
		return nil
	}
	var p []*Person
	for _, mp := range mps {
		p = append(p, mp.toPerson())
	}

	return p
}

func (s *MongoStorage) GetPersonsByCommunication(value string) []*Person {
	collection := s.client.Database(dbName).Collection(collectionName)

	cursor, _ := collection.Find(s.ctx, bson.D{{"communication.value", value}})

	var mps []*MongoPerson
	err := cursor.All(s.ctx, &mps)
	if err != nil || len(mps) == 0 {
		return nil
	}
	var p []*Person
	for _, mp := range mps {
		p = append(p, mp.toPerson())
	}

	return p
}

func (s *MongoStorage) UpdatePerson(person *Person) bool {
	p := s.GetPersonById(person.ID)
	if p == nil {
		return false
	}

	collection := s.client.Database(dbName).Collection(collectionName)
	mp := person.toMongoPerson()
	_, err := collection.ReplaceOne(s.ctx, bson.D{{"_id", mp.ID}}, mp)
	if err != nil {
		return false
	}
	return true
}

func (s *MongoStorage) DeletePerson(id uuid.UUID) bool {
	p := s.GetPersonById(id)
	if p == nil {
		return false
	}

	collection := s.client.Database(dbName).Collection(collectionName)
	_, err := collection.DeleteOne(s.ctx, bson.D{{"_id", id.String()}})
	if err != nil {
		return false
	}
	return true
}

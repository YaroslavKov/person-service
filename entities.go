package main

import (
	uuid "github.com/satori/go.uuid"
)

type Person struct {
	ID             uuid.UUID        `json:"id"`
	Name           string           `json:"name"`
	Communications []*Communication `json:"communications"`
}

type Communication struct {
	Value string `json:"value"`
}

type MongoPerson struct {
	ID             string           `bson:"_id"`
	Name           string           `bson:"name"`
	Communications []*Communication `bson:"communication"`
}

func (p *MongoPerson) toPerson() *Person {
	return &Person{
		ID:             uuid.FromStringOrNil(p.ID),
		Name:           p.Name,
		Communications: p.Communications,
	}
}

func (p *Person) toMongoPerson() *MongoPerson {
	return &MongoPerson{
		ID:             p.ID.String(),
		Name:           p.Name,
		Communications: p.Communications,
	}
}

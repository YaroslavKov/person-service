package main

import uuid "github.com/satori/go.uuid"

type InMemoryPersonStorage struct {
	data map[uuid.UUID]*Person
}

func NewInMemoryPersonStorage() *InMemoryPersonStorage {
	return &InMemoryPersonStorage{make(map[uuid.UUID]*Person)}
}

func (s *InMemoryPersonStorage) Add(person *Person) error {
	_, ok := s.data[person.ID]
	if ok {
		return personExistError
	}

	s.data[person.ID] = person
	return nil
}

func (s *InMemoryPersonStorage) GetAll() []*Person {
	persons := make([]*Person, 0)
	for _, person := range s.data {
		persons = append(persons, person)
	}
	return persons
}

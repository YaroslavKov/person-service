package main

import uuid "github.com/satori/go.uuid"

type InMemoryPersonStorage struct {
	data map[uuid.UUID]*Person
}

func NewInMemoryPersonStorage() *InMemoryPersonStorage {
	return &InMemoryPersonStorage{make(map[uuid.UUID]*Person)}
}

func (s *InMemoryPersonStorage) Add(person *Person) error {
	p := s.GetPersonById(person.ID)
	if p != nil {
		return personExistError
	}

	s.data[person.ID] = person
	return nil
}

func (s *InMemoryPersonStorage) GetAll() []*Person {
	var persons []*Person
	for _, person := range s.data {
		persons = append(persons, person)
	}
	return persons
}

func (s *InMemoryPersonStorage) GetPersonById(id uuid.UUID) *Person {
	p, ok := s.data[id]
	if !ok {
		return nil
	}
	return p
}

func (s *InMemoryPersonStorage) GetPersonsByName(name string) []*Person {
	var persons []*Person
	for _, val := range s.data {
		if val.Name == name {
			persons = append(persons, val)
		}
	}
	return persons
}

func (s *InMemoryPersonStorage) GetPersonsByCommunication(value string) []*Person {
	var persons []*Person
	for _, val := range s.data {
		for _, comm := range val.Communications {
			if comm.Value == value {
				persons = append(persons, val)
				break
			}
		}
	}
	return persons
}

func (s *InMemoryPersonStorage) UpdatePerson(person *Person) bool {
	p := s.GetPersonById(person.ID)
	if p == nil {
		return false
	}
	s.data[person.ID] = person
	return true
}

func (s *InMemoryPersonStorage) DeletePerson(id uuid.UUID) bool {
	p := s.GetPersonById(id)
	if p == nil {
		return false
	}
	delete(s.data, id)
	return true
}

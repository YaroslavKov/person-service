package main

import uuid "github.com/satori/go.uuid"

type InMemoryPersonStorage struct {
	data map[uuid.UUID]*Person
}

func NewInMemoryPersonStorage() *InMemoryPersonStorage {
	return &InMemoryPersonStorage{make(map[uuid.UUID]*Person)}
}

func (s *InMemoryPersonStorage) GetAll() ([]*Person, error) {
	persons := []*Person{}
	for _, person := range s.data {
		persons = append(persons, person)
	}
	return persons, nil
}

func (s *InMemoryPersonStorage) Add(person *Person) (*Person, error) {
	p, err := s.GetPersonByID(person.ID)
	if err == nil {
		return p, personExistError
	}

	s.data[person.ID] = person
	return p, nil
}

func (s *InMemoryPersonStorage) GetPersonByID(id uuid.UUID) (*Person, error) {
	p, ok := s.data[id]
	if !ok {
		return nil, personNotFoundError
	}
	return p, nil
}

func (s *InMemoryPersonStorage) GetPersonsByName(name string) ([]*Person, error) {
	persons := []*Person{}
	for _, val := range s.data {
		if val.Name == name {
			persons = append(persons, val)
		}
	}
	if len(persons) == 0 {
		return nil, personNotFoundError
	}
	return persons, nil
}

func (s *InMemoryPersonStorage) GetPersonsByCommunication(value string) ([]*Person, error) {
	persons := []*Person{}
	for _, val := range s.data {
		for _, comm := range val.Communications {
			if comm.Value == value {
				persons = append(persons, val)
				break
			}
		}
	}
	if len(persons) == 0 {
		return nil, personNotFoundError
	}
	return persons, nil
}

func (s *InMemoryPersonStorage) UpdatePerson(person *Person) (*Person, error) {
	_, err := s.GetPersonByID(person.ID)
	if err != nil {
		return nil, err
	}
	s.data[person.ID] = person
	return person, nil
}

func (s *InMemoryPersonStorage) DeletePerson(id uuid.UUID) (*Person, error) {
	p, err := s.GetPersonByID(id)
	if err != nil {
		return nil, err
	}
	delete(s.data, id)
	return p, nil
}

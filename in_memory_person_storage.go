package main

type InMemoryPersonStorage struct {
	data map[string]Person
}

func NewInMemoryPersonStorage() (storage *InMemoryPersonStorage) {
	return &InMemoryPersonStorage{make(map[string]Person)}
}

func (storage *InMemoryPersonStorage) Add(person Person) error {
	_, ok := storage.data[person.ID]
	if ok {
		return PersonExistError
	}

	//add ID validation for uuid

	storage.data[person.ID] = person
	return nil
}

func (storage *InMemoryPersonStorage) GetAll() (persons []Person) {
	for _, person := range storage.data {
		persons = append(persons, person)
	}
	return
}

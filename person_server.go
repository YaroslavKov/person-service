package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type PersonServer struct {
	storage PersonStorage
	http.Handler
}

type PersonStorage interface {
	GetAll() []Person
	Add(Person) error
}

func NewPersonServer(storage PersonStorage) (server *PersonServer) {
	server = new(PersonServer)
	server.storage = storage

	mux := http.NewServeMux()
	mux.Handle("/person", http.HandlerFunc(server.personHandler))

	server.Handler = mux

	return
}

func (server *PersonServer) personHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		server.addPerson(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (server *PersonServer) addPerson(w http.ResponseWriter, r *http.Request) {
	person, err := decodeBody(r.Body)
	if err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	err = server.storage.Add(person)
	if err == PersonExistError {
		handleError(err, w, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func decodeBody(body io.Reader) (person Person, err error) {
	err = json.NewDecoder(body).Decode(&person)
	if err == io.EOF {
		err = nil
	}
	return
}

func handleError(err error, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}

package main

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	storage Storage
	http.Handler
}

type Storage interface {
	GetAll() []*Person
	Add(*Person) error
	//GetById(uuid uuid.UUID) *Person
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(storage Storage) *Server {
	server := &Server{storage: storage}

	mux := http.NewServeMux()
	mux.Handle("/person", http.HandlerFunc(server.personHandler))

	server.Handler = mux

	return server
}

func (s *Server) personHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.addPerson(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) addPerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeJSON)

	if err := isContentTypeJson(r); err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	p := &Person{}
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	if err := s.storage.Add(p); err == personExistError {
		handleError(err, w, http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.storage.GetAll())
}

func handleError(err error, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
}

func isContentTypeJson(r *http.Request) error {
	if r.Header.Get("Content-Type") != contentTypeJSON {
		return wrongContentType
	}
	return nil
}

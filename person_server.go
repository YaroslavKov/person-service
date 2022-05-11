package main

import (
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
)

type Server struct {
	storage Storage
	http.Handler
}

type Storage interface {
	GetAll() []*Person
	Add(*Person) error
	GetPersonById(uuid.UUID) *Person
	GetPersonsByName(string) []*Person
	GetPersonsByCommunication(string) []*Person
	UpdatePerson(*Person) bool
	DeletePerson(uuid.UUID) bool
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(storage Storage) *Server {
	server := &Server{storage: storage}

	mux := http.NewServeMux()
	personHandler := server.requestAuthentication(server.personHandler)
	mux.Handle("/person", personHandler)
	mux.Handle("/person/", personHandler)

	server.Handler = mux

	return server
}

func (s *Server) requestAuthentication(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || username != authLogin || password != authPassword {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next(w, r)
		}
	})
}

func (s *Server) personHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()

	if !ok || username != authLogin || password != authPassword {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)

	switch r.Method {
	case http.MethodGet:
		s.getPersons(w, r)
	case http.MethodPost:
		s.addPerson(w, r)
	case http.MethodPut:
		s.putPerson(w, r)
	case http.MethodDelete:
		s.deletePerson(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) addPerson(w http.ResponseWriter, r *http.Request) {
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
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (s *Server) getPersons(w http.ResponseWriter, r *http.Request) {
	if idStr := getQueryParam(r, "id"); idStr != "" {
		id, err := uuid.FromString(idStr)
		if err != nil {
			handleError(err, w, http.StatusBadRequest)
			return
		}

		p := s.storage.GetPersonById(id)
		if p == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(p)
		w.WriteHeader(http.StatusOK)
		return
	}

	if name := getQueryParam(r, "name"); name != "" {
		p := s.storage.GetPersonsByName(name)
		if p == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(p)
		w.WriteHeader(http.StatusOK)
		return
	}

	if communication := getQueryParam(r, "communication"); communication != "" {
		p := s.storage.GetPersonsByCommunication(communication)
		if p == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(p)
		w.WriteHeader(http.StatusOK)
		return
	}

	handleError(errors.New("use id, name or communication param"), w, http.StatusBadRequest)
}

func (s *Server) putPerson(w http.ResponseWriter, r *http.Request) {
	if err := isContentTypeJson(r); err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	p := &Person{}
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(p)
	if ok := s.storage.UpdatePerson(p); !ok {
		s.storage.Add(p)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deletePerson(w http.ResponseWriter, r *http.Request) {
	if idStr := strings.TrimPrefix(r.URL.Path, "/person/"); idStr != "" {
		id, err := uuid.FromString(idStr)
		if err != nil {
			handleError(err, w, http.StatusBadRequest)
			return
		}

		if ok := s.storage.DeletePerson(id); ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		return
	}

	handleError(invalidUuidError, w, http.StatusBadRequest)
}

func handleError(err error, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
}

func isContentTypeJson(r *http.Request) error {
	if r.Header.Get("Content-Type") != contentTypeJSON {
		return wrongContentTypeError
	}
	return nil
}

func getQueryParam(r *http.Request, paramName string) string {
	if keys, ok := r.URL.Query()[paramName]; ok && len(keys[0]) > 0 {
		return keys[0]
	}
	return ""
}

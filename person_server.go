package main

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
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
	personHandler := server.requestAuthentication(server.logging(http.HandlerFunc(server.personHandler)))
	mux.Handle("/person", personHandler)
	mux.Handle("/person/", personHandler)

	server.Handler = mux

	return server
}

func (s *Server) requestAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok || username != authLogin || password != authPassword {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, f := getServiceLogger("person_server")
		if logger == nil {
			next.ServeHTTP(w, r)
			return
		}
		defer f.Close()

		someFlag := true

		var buf []byte
		if r.Body != nil {
			buf, _ = ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		}

		f.WriteString("\n")
		e := logger.Info().Time("time", time.Now()).Bytes("method", []byte(r.Method)).Bytes("path", []byte(r.URL.Path)).Bytes("agent", []byte(r.Header.Get("User-Agent")))
		if r.Body != nil && isContentTypeJson(r) && someFlag {
			e.Bytes("body", buf)
		}
		e.Send()

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		e = logger.Info().Time("time", time.Now()).Int("status", lrw.statusCode)
		if r.Body != nil && someFlag {
			e.Bytes("body", lrw.body)
		}
		e.Send()
	})
}

func (s *Server) personHandler(w http.ResponseWriter, r *http.Request) {
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
	if !isContentTypeJson(r) {
		handleError(wrongContentTypeError, w, http.StatusBadRequest)
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

	ps := s.storage.GetAll()
	if ps == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ps)
	w.WriteHeader(http.StatusOK)
	return
}

func (s *Server) putPerson(w http.ResponseWriter, r *http.Request) {
	if !isContentTypeJson(r) {
		handleError(wrongContentTypeError, w, http.StatusBadRequest)
		return
	}

	p := &Person{}
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	if ok := s.storage.UpdatePerson(p); !ok {
		s.storage.Add(p)
		json.NewEncoder(w).Encode(s.storage.GetPersonById(p.ID))
		w.WriteHeader(http.StatusOK)
		return
	}
	json.NewEncoder(w).Encode(s.storage.GetPersonById(p.ID))
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

func isContentTypeJson(r *http.Request) bool {
	if r.Header.Get("Content-Type") == contentTypeJSON {
		return true
	}
	return false
}

func getQueryParam(r *http.Request, paramName string) string {
	if keys, ok := r.URL.Query()[paramName]; ok && len(keys[0]) > 0 {
		return keys[0]
	}
	return ""
}

func getServiceLogger(name string) (*zerolog.Logger, *os.File) {
	const folder = "logs"

	if err := os.Mkdir(folder, os.FileMode(0755)); err != nil && !os.IsExist(err) {
		return nil, nil
	}

	f, err := os.OpenFile(folder+"/"+name+".log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil
	}
	mainLogger := zerolog.New(f).With().Logger()
	return &mainLogger, f
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, nil}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body = append(lrw.body, b...)
	return lrw.ResponseWriter.Write(b)
}

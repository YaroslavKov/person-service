package main

import (
	"bytes"
	"encoding/json"
	"github.com/go-http-utils/headers"
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
	logBody bool
}

type Storage interface {
	GetAll() ([]*Person, error)
	Add(*Person) (*Person, error)
	GetPersonByID(uuid.UUID) (*Person, error)
	GetPersonsByName(string) ([]*Person, error)
	GetPersonsByCommunication(string) ([]*Person, error)
	UpdatePerson(*Person) (*Person, error)
	DeletePerson(uuid.UUID) (*Person, error)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(storage Storage, logBody bool) *Server {
	server := &Server{storage: storage, logBody: logBody}

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

		var buf []byte
		if r.Body != nil {
			buf, _ = ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		}

		f.WriteString("\n")
		e := logger.Info().Time("time", time.Now()).Bytes("method", []byte(r.Method)).Bytes("path", []byte(r.URL.Path)).Bytes("agent", []byte(r.Header.Get("User-Agent")))
		if r.Body != nil && isContentTypeJSON(r) && s.logBody {
			e.Bytes("body", buf)
		}
		e.Send()

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		e = logger.Info().Time("time", time.Now()).Int("status", lrw.statusCode)
		if lrw.body != nil && s.logBody {
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
	if !isContentTypeJSON(r) {
		handleError(wrongContentTypeError, w, http.StatusUnsupportedMediaType)
		return
	}

	p := &Person{}
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	addedPerson, err := s.storage.Add(p)
	if err == personExistError {
		handleError(err, w, http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		handleError(err, w, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(addedPerson)
}

func (s *Server) getPersons(w http.ResponseWriter, r *http.Request) {
	if idStr := strings.TrimPrefix(r.URL.Path, "/person/"); strings.HasPrefix(r.URL.Path, "/person/") && idStr != "" {
		id, err := uuid.FromString(idStr)
		if err != nil {
			handleError(err, w, http.StatusBadRequest)
			return
		}
		p, err := s.storage.GetPersonByID(id)
		if err == personNotFoundError {
			handleError(err, w, http.StatusNotFound)
			return
		} else if err != nil {
			handleError(err, nil, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(p)
		w.WriteHeader(http.StatusOK)
		return
	}

	ppName := []*Person{}
	ppComm := []*Person{}

	if name := getQueryParam(r, "name"); name != "" {
		pName, err := s.storage.GetPersonsByName(name)
		if err == nil {
			ppName = append(ppName, pName...)
		} else if err != personNotFoundError {
			handleError(err, w, http.StatusInternalServerError)
			return
		}
	}
	if communication := getQueryParam(r, "communication"); communication != "" {
		pName, err := s.storage.GetPersonsByCommunication(communication)
		if err == nil {
			ppComm = append(ppComm, pName...)
		} else if err != personNotFoundError {
			handleError(err, w, http.StatusInternalServerError)
			return
		}
	}

	if len(r.URL.Query()) != 0 {
		var pp []*Person
		if len(ppName) != 0 && len(ppComm) != 0 {
			pp = exceptPersons(ppName, ppComm)
		} else if len(ppName) != 0 {
			pp = ppName
		} else if len(ppComm) != 0 {
			pp = ppComm
		}
		if len(pp) == 0 {
			handleError(personNotFoundError, w, http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(pp)
		w.WriteHeader(http.StatusOK)
		return
	}

	ps, err := s.storage.GetAll()
	if err != nil {
		handleError(err, w, http.StatusInternalServerError)
		return
	} else if len(ps) == 0 {
		handleError(personNotFoundError, w, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ps)
	w.WriteHeader(http.StatusOK)
	return
}

func (s *Server) putPerson(w http.ResponseWriter, r *http.Request) {
	if !isContentTypeJSON(r) {
		handleError(wrongContentTypeError, w, http.StatusUnsupportedMediaType)
		return
	}

	p := &Person{}
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		handleError(err, w, http.StatusBadRequest)
		return
	}

	p2, err := s.storage.UpdatePerson(p)
	if err == personNotFoundError {
		p2, err = s.storage.Add(p)
		json.NewEncoder(w).Encode(p2)
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		handleError(err, w, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(p2)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deletePerson(w http.ResponseWriter, r *http.Request) {
	if idStr := strings.TrimPrefix(r.URL.Path, "/person/"); idStr != "" {
		id, err := uuid.FromString(idStr)
		if err != nil {
			handleError(err, w, http.StatusBadRequest)
			return
		}

		_, err = s.storage.DeletePerson(id)
		if err == personNotFoundError {
			handleError(err, w, http.StatusNotFound)
		} else if err != nil {
			handleError(err, w, http.StatusInternalServerError)
		} else if err == nil {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	handleError(invalidUuidError, w, http.StatusBadRequest)
}

func handleError(err error, w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
}

func isContentTypeJSON(r *http.Request) bool {
	if r.Header.Get(headers.ContentType) == contentTypeJSON {
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

func exceptPersons(p1 []*Person, p2 []*Person) []*Person {
	pRes := []*Person{}
	for _, value1 := range p1 {
		for _, value2 := range p2 {
			if value1.ID == value2.ID {
				pRes = append(pRes, value1)
				break
			}
		}
	}
	return pRes
}

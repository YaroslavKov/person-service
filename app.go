package main

import (
	"log"
	"net/http"
)

const port = ":5002"

func main() {
	var server *Server
	switch 2 {
	case 1:
		storage, err := NewMongoStorage()
		if err != nil {
			log.Panic(err)
		}
		defer storage.Close()
		server = NewServer(storage)
	case 2:
		storage, err := NewPostgresStorage()
		if err != nil {
			log.Panic(err)
		}
		server = NewServer(storage)
	default:
		storage := NewInMemoryPersonStorage()
		server = NewServer(storage)
	}

	if err := http.ListenAndServe(port, server); err != nil {
		log.Fatalf("could not listen on port %v %v", port, err)
	}
}

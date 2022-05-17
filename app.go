package main

import (
	"log"
	"net/http"
)

const port = ":5002"

func main() {
	var server *Server
	switch 1 {
	case 1:
		storage, _ := NewMongoStorage()
		defer storage.Close()
		server = NewServer(storage)
	default:
		storage := NewInMemoryPersonStorage()
		server = NewServer(storage)
	}

	if err := http.ListenAndServe(port, server); err != nil {
		log.Fatalf("could not listen on port %v %v", port, err)
	}
}

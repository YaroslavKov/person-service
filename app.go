package main

import (
	"log"
	"net/http"
	"os"
)

const port = ":5002"

func main() {
	logBody := false
	storage := ""

	args := os.Args
	if len(args) >= 2 {
		for i, arg := range args {
			if arg == "-logBody" {
				logBody = true
			}

			if arg == "--storage" || arg == "-s" {
				if i < len(args)-1 {
					storage = args[i+1]
				}
			}
		}
	}

	var server *Server
	switch storage {
	case "mongo":
		storage, err := NewMongoStorage()
		if err != nil {
			log.Panic(err)
		}
		defer storage.Close()
		server = NewServer(storage, logBody)
	case "postgres":
		storage, err := NewPostgresStorage()
		if err != nil {
			log.Panic(err)
		}
		server = NewServer(storage, logBody)
	default:
		storage := NewInMemoryPersonStorage()
		server = NewServer(storage, logBody)
	}

	if err := http.ListenAndServe(port, server); err != nil {
		log.Fatalf("could not listen on port %v %v", port, err)
	}
}

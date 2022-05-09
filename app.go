package main

import (
	"log"
	"net/http"
)

const port = ":5002"

func main() {

	var server http.Handler = NewServer(NewInMemoryPersonStorage())

	if err := http.ListenAndServe(port, server); err != nil {
		log.Fatalf("could not listen on port %v %v", port, err)
	}
}

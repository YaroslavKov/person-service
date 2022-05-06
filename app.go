package main

import (
	"log"
	"net/http"
)

const port = ":5002"

func main() {

	var server http.Handler = NewPersonServer(NewInMemoryPersonStorage())

	if err := http.ListenAndServe(port, server); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
	//log.Printf("Listening on port %v\n", port)
}

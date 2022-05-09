package main

import uuid "github.com/satori/go.uuid"

type Person struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Communications []Communication `json:"communications"`
}

type Communication struct {
	Value string `json:"value"`
}

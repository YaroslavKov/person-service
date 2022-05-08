package main

import uuid "github.com/satori/go.uuid"

type Person struct {
	ID             uuid.UUID
	Name           string
	Communications []Communication
}

type Communication struct {
	Value string
}

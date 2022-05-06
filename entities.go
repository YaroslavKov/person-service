package main

type Person struct {
	ID             string
	Name           string
	Communications []Communication
}

type Communication struct {
	Value string
}

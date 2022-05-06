package main

import "errors"

var (
	PersonExistError    = errors.New("Person already exist")
	NotValidPersonError = errors.New("Person not valid")
)

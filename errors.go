package main

import "errors"

var (
	personExistError    = errors.New("person already exist")
	notValidPersonError = errors.New("person not valid")
	wrongContentType    = errors.New("wrong content type")
)

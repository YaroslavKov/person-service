package main

import "errors"

var (
	personExistError      = errors.New("person already exist")
	notValidPersonError   = errors.New("person not valid")
	wrongContentTypeError = errors.New("wrong content type")
	invalidUuidError      = errors.New("invalid uuid")
)

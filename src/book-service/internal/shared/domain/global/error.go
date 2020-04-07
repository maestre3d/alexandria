package global

import (
	"errors"
)

// InvalidID Entity has an invalid ID
var InvalidID = errors.New("invalid id")

// RequiredField A field is required
var RequiredField = errors.New("missing required request field")

// InvalidFieldFormat A field has a bad format
var InvalidFieldFormat = errors.New("request field %s has an invalid format, expected %s")

// EntityDomainError Something happened at domain.entity
var EntityDomainError = errors.New("entity domain failure")

// EmptyQuery Query is empty
var EmptyQuery = errors.New("empty query")

// EntityNotFound Entity was not found
var EntityNotFound = errors.New("resource not found")

// EntitiesNotFound Entities were not found
var EntitiesNotFound = errors.New("resources not found")

// EntityExists Entity already exists
var EntityExists = errors.New("resource already exists")

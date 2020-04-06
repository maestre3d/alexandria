package global

import (
	"errors"
)

// InvalidID Entity has an invalid ID
var InvalidID = errors.New("invalid ID")

// RequiredField A field is required
var RequiredField = errors.New("missing required field")

// InvalidFieldFormat A field has a bad format
var InvalidFieldFormat = errors.New("invalid field format")

// EntityDomainError Something happened at domain.entity
var EntityDomainError = errors.New("entity domain failure")

// EmptyQuery Query is empty
var EmptyQuery = errors.New("empty query")

// EntityNotFound Entity was not found
var EntityNotFound = errors.New("entity not found")

// EntitiesNotFound Entities were not found
var EntitiesNotFound = errors.New("entities not found")

// EntityExists Entity already exists
var EntityExists = errors.New("entity already exists")

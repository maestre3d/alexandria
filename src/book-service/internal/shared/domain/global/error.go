package global

import "errors"

// InvalidID Entity has an invalid ID
var InvalidID = errors.New("invalid ID")

// EmptyQuery Query is empty
var EmptyQuery = errors.New("empty query")

// EntityNotFound Entity was not found
var EntityNotFound = errors.New("entity not found")

// EntityExists Entity already exists
var EntityExists = errors.New("entity already exists")

package domain

import (
	"errors"
	"github.com/alexandria-oss/core/exception"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAuthor(t *testing.T) {
	author := NewAuthor("Isaac", "Newton", "", "private", NewOwner(
		"1b4cc750-c551-4767-a232-e91b52e68fa0",
		"admin"))
	if err := author.IsValid(); err != nil {
		assert.Equal(t, true, errors.Is(err, exception.RequiredField))
	}

	assert.Equal(t, nil, author.IsValid())

	author.Owners = append(author.Owners, NewOwner(
		"69817804-4af4-4de1-83af-4a5f660d0018",
		"ContriB"))

	assert.Equal(t, nil, author.IsValid())

	a2 := NewAuthor("Aex", "12", "", "public", nil)
	if err := a2.IsValid(); err != nil {
		assert.Equal(t, true, errors.Is(err, exception.InvalidFieldRange))
	}

	a2.Owners = append(a2.Owners, NewOwner(
		"69817804-4af4-4de1-83af-4a5f660d0018",
		"owner"))

	assert.Equal(t, nil, a2.IsValid())
}

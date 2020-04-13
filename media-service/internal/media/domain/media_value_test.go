package domain

import (
	"errors"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/global"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMediaID_IsValid(t *testing.T) {
	// Sending with negative id
	invalidID := mediaID{-5}
	err := invalidID.IsValid() // should return invalid ID error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidID))

	// Sending 0-value
	zeroID := mediaID{0}
	err = zeroID.IsValid() // should return invalid ID error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidID))

	// Sending valid value
	id := mediaID{9999}
	err = id.IsValid() // Should be valid
	assert.Nil(t, err)
}

func TestExternalID_Generate(t *testing.T) {
	// Check correct UUID generating
	eID := externalID{""}
	eID.Generate()
	_, err := uuid.Parse(eID.Value)
	assert.Nil(t, err)
}

func TestExternalID_IsValid(t *testing.T) {
	// Sending no data
	emptyExtID := externalID{""}
	err := emptyExtID.IsValid() // Should return err, RequiredField
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.RequiredField))

	// Sending invalid UUID
	invalidExtID := externalID{"cb031d46-gab6-4dbc-9a6d-1e123206ff7g"}
	err = invalidExtID.IsValid()
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldFormat))

	// Sending valid UUID
	validUUID := externalID{"cb031d46-cab6-4dbc-9a6d-1e123206ff7b"}
	err = validUUID.IsValid()
	assert.Nil(t, err)
}

func TestTitle_IsValid(t *testing.T) {
	// Sending empty title
	titleTest := title{""}
	err := titleTest.IsValid() // Should return requiredField error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.RequiredField))

	// Sending out of range title
	for i := 1; i <= 256; i++ {
		titleTest.Value += "x"
	}
	// Value should be 256 times x
	err = titleTest.IsValid() // Should return invalidFieldRange error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldRange))

	// Sending valid title
	titleTest.Value = ""
	for i := 1; i <= 255; i++ {
		titleTest.Value += "x"
	}
	err = titleTest.IsValid()
	assert.Nil(t, err)
}

func TestDisplayName_IsValid(t *testing.T) {
	// Send empty display_name
	displayTest := displayName{""}
	err := displayTest.IsValid() // Should return requiredField error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.RequiredField))

	// Sending invalid display_name
	for i := 1; i <= 101; i++ {
		displayTest.Value += "x"
	}
	// Should be 101 times x
	err = displayTest.IsValid() // Should return invalidFieldRange error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldRange))

	// Sending valid display_name
	displayTest.Value = ""
	for i := 1; i <= 100; i++ {
		displayTest.Value += "x"
	}
	err = displayTest.IsValid()
	assert.Nil(t, err)
}

func TestDescription_IsValid(t *testing.T) {
	// Send empty description
	desc := description{nil}
	err := desc.IsValid() // Should be valid
	assert.Nil(t, err)

	// Sending out of range value
	strObj := ""
	for i := 1; i <= 1025; i++ {
		strObj += "x"
	}
	desc.Value = &strObj
	err = desc.IsValid() // Should return invalidFieldRange error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldRange))

	// Sending valid, not empty description
	strObj = ""
	for i := 1; i <= 1024; i++ {
		strObj += "x"
	}
	desc.Value = &strObj
	err = desc.IsValid()
	assert.Nil(t, err)
}

func TestUserID_IsValid(t *testing.T) {
	// Send empty userID
	userID := UserID{""}
	err := userID.IsValid() // Should return requiredField error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.RequiredField))

	// Send invalid UUID
	userID.Value = "9f001628-fd97-4a8a-b2db-9d8972fc3d7g"
	err = userID.IsValid() // Should return invalidFieldFormat
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldFormat))

	// Send valid UUID
	userID.Value = "9f001628-fd97-4a8a-b2db-9d8972fc3d75"
	err = userID.IsValid()
	assert.Nil(t, err)
}

func TestAuthorID_IsValid(t *testing.T) {
	// Send empty userID
	authorID := AuthorID{""}
	err := authorID.IsValid() // Should return requiredField error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.RequiredField))

	// Send invalid UUID
	authorID.Value = "9f001628-fd97-4a8a-b2db-9d8972fc3d7g"
	err = authorID.IsValid() // Should return invalidFieldFormat
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldFormat))

	// Send valid UUID
	authorID.Value = "9f001628-fd97-4a8a-b2db-9d8972fc3d75"
	err = authorID.IsValid()
	assert.Nil(t, err)
}

func TestMediaType_IsValid(t *testing.T) {
	// Send empty media type
	mediaType := MediaType{""}
	err := mediaType.IsValid() // Should return requiredField error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.RequiredField))

	// Send invalid media type
	mediaType.Value = "MEDIA_BOOKS"
	err = mediaType.IsValid() // Should return invalidFieldFormat error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldFormat))

	// It mustn't accept lowercase
	mediaType.Value = "media_book"
	err = mediaType.IsValid() // Should return invalidFieldFormat error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldFormat))

	// Send valid media type
	mediaType.Value = "MEDIA_BOOK"
	err = mediaType.IsValid()
	assert.Nil(t, err)
}

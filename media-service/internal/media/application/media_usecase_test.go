package application

import (
	"errors"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/media/infrastructure"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

type MockLogger struct {
	mock.Mock
}

var logger = new(MockLogger)
var repository = infrastructure.NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), logger)
var usecase = NewMediaUseCase(logger, repository)
var params = &MediaParams{
	MediaID:     "",
	Title:       "The foo programing language",
	DisplayName: "Foo book",
	Description: "The foo programming language has become one of the top notch...",
	UserID:      "60f90323-fc78-45e4-a0f5-71b63dd87d1a",
	AuthorID:    "a38d10fa-f369-4e8c-8c9d-f7f9f22bdc71",
	PublishDate: "2006-12-31",
	MediaType:   "media_book",
}

func TestMediaUseCase_Create(t *testing.T) {
	// Send nil media
	err := usecase.Create(nil) // Should return requiredField error
	assert.NotNil(t, err)

	// Send invalid media
	params.UserID = "60f90323-fc78-45e4-a0f5-71b63dd87d1g"
	params.PublishDate = "2006-31-12"
	params.MediaType = "media_booKs"
	err = usecase.Create(params) // Should return invalidFieldFormat error
	assert.NotNil(t, err)

	// Send valid data
	params.Title = "The foo programing language"
	params.UserID = "60f90323-fc78-45e4-a0f5-71b63dd87d1a"
	params.MediaType = "media_book"
	params.PublishDate = "2006-12-31"
	err = usecase.Create(params)
	assert.Nil(t, err)
}

func TestMediaUseCase_GetAll(t *testing.T) {
	// Add data
	err := usecase.Create(params)
	assert.Nil(t, err)

	// Send nil paginateParams and filters
	medias, err := usecase.GetAll(nil, nil) // Should be valid
	assert.Nil(t, err)
	assert.True(t, len(medias) >= 1)

	// Send invalid paginateParams
	paginateParams := &util.PaginationParams{
		TokenID:   0,
		TokenUUID: "qwerty",
		Size:      0,
	}
	_, err = usecase.GetAll(paginateParams, nil) // Should be valid
	assert.Nil(t, err)

	// Send valid page_token and invalid page_size
	paginateParams.TokenUUID = medias[0].ExternalID
	paginateParams.Size = 999
	_, err = usecase.GetAll(paginateParams, nil)
	assert.Nil(t, err)
}

func TestMediaUseCase_GetByID(t *testing.T) {
	// Add data
	err := usecase.Create(params)
	assert.Nil(t, err)

	// Get
	medias, err := usecase.GetAll(nil, nil) // Should return previous created media
	assert.Nil(t, err)
	assert.True(t, len(medias) >= 1)

	// Send invalid UUID
	_, err = usecase.GetByID("qwerty") // Should return invalidID error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidID))

	// Send valid ID (not UUID)
	_, err = usecase.GetByID("1") // Should be valid
	assert.Nil(t, err)

	// Send valid UUID
	media, err := usecase.GetByID(medias[0].ExternalID) // Should return the previous created media object
	assert.Nil(t, err)
	assert.Equal(t, medias[0].ExternalID, media.ExternalID)
}

func TestMediaUseCase_GetByTitle(t *testing.T) {
	// Add data
	err := usecase.Create(params)
	assert.Nil(t, err)

	// Send invalid name
	_, err = usecase.GetByTitle("qwerty") // Should return entityNotFound
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.EntityNotFound))

	// Send valid existing uppercase title
	media, err := usecase.GetByTitle(strings.ToUpper(params.Title)) // Should return previous created media
	assert.Nil(t, err)
	assert.Equal(t, media.Title, params.Title)
}

func TestMediaUseCase_UpdateOneAtomic(t *testing.T) {
	// Add data
	err := usecase.Create(params)
	assert.Nil(t, err)

	// Verify media was added
	medias, err := usecase.GetAll(nil, nil)
	assert.Nil(t, err)
	assert.True(t, len(medias) >= 1)

	// Send empty
	err = usecase.UpdateOne(nil) // Should return emptyBody error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.EmptyBody))

	// Send non-atomic params
	paramUpdate := &MediaParams{
		MediaType: "media_books",
	}
	err = usecase.UpdateOneAtomic(paramUpdate) // Should return error
	assert.NotNil(t, err)

	// Send invalid no ID/UUID
	err = usecase.UpdateOne(paramUpdate) // Should return invalidID error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidID))

	// Send invalid field
	params.MediaID = medias[0].ExternalID
	params.MediaType = "media_books"
	err = usecase.UpdateOneAtomic(params) // Should return invalidFieldFormat error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldFormat))

	// Send valid field
	params.MediaType = "media_doc"
	err = usecase.UpdateOneAtomic(params)
	assert.Nil(t, err)
}

func TestMediaUseCase_UpdateOne(t *testing.T) {
	// Add data
	err := usecase.Create(params)
	assert.Nil(t, err)

	// Verify media was added
	medias, err := usecase.GetAll(nil, nil)
	assert.Nil(t, err)
	assert.True(t, len(medias) >= 1)

	// Send empty
	err = usecase.UpdateOne(nil) // Should return emptyBody error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.EmptyBody))

	paramUpdate := &MediaParams{
		MediaType: "media_books",
	}

	// Send invalid no ID/UUID
	err = usecase.UpdateOne(paramUpdate) // Should return invalidID error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidID))

	// Send invalid field
	paramUpdate.MediaID = medias[0].ExternalID
	err = usecase.UpdateOne(paramUpdate) // Should return invalidFieldFormat error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidFieldFormat))

	// Send valid data
	paramUpdate.MediaType = "media_book"
	err = usecase.UpdateOne(paramUpdate)
	assert.Nil(t, err)
}

func TestMediaUseCase_RemoveOne(t *testing.T) {
	// Add data
	err := usecase.Create(params)
	assert.Nil(t, err)

	// Verify media was added
	medias, err := usecase.GetAll(nil, nil)
	assert.Nil(t, err)
	assert.True(t, len(medias) >= 1)

	// Send invalid UUID
	err = usecase.RemoveOne("qwerty") // Should return invalidID error
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.InvalidID))

	// Send valid ID
	err = usecase.RemoveOne(medias[0].ExternalID)
	assert.Nil(t, err)
}

// Mock logger
func (m *MockLogger) Print(message, resource string) {
	m.Called(message, resource)
}

func (m *MockLogger) Error(message, resource string) {
	m.Called(message, resource)
}

func (m *MockLogger) Fatal(message, resource string) {
	m.Called(message, resource)
}

func (m *MockLogger) Close() func() {
	m.Called()
	return func() {}
}

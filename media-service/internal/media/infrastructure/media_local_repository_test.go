package infrastructure

import (
	"errors"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockLogger struct {
	mock.Mock
}

var mediaParams = &domain.MediaEntityParams{
	Title:       "The foo programming language",
	DisplayName: "Foo book",
	Description: "The foo programming language has become one of the top notch...",
	UserID:      "60f90323-fc78-45e4-a0f5-71b63dd87d1a",
	AuthorID:    "a38d10fa-f369-4e8c-8c9d-f7f9f22bdc71",
	PublishDate: "2006-12-31",
	MediaType:   "media_book",
}

func TestMediaLocalRepository_Save(t *testing.T) {
	// Send valid media
	localRepo := NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), new(MockLogger))
	media, err := domain.NewMediaEntity(mediaParams)
	assert.Nil(t, err)
	assert.Nil(t, localRepo.Save(media.ToMediaAggregate()))

	// Send replicated data
	assert.NotNil(t, localRepo.Save(media.ToMediaAggregate())) // Should fail
}

func TestMediaLocalRepository_Fetch(t *testing.T) {
	// Store a multiple media
	localRepo := NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), new(MockLogger))

	_, err := localRepo.Fetch(nil, nil)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, global.EntitiesNotFound))

	mediaParam2 := &domain.MediaEntityParams{
		Title:       "The bar programming language",
		DisplayName: "Bar book",
		Description: "The bar programming language has become one of the top notch...",
		UserID:      "60f90323-fc78-45e4-a0f5-71b63dd87d1a",
		AuthorID:    "a38d10fa-f369-4e8c-8c9d-f7f9f22bdc71",
		PublishDate: "2006-12-31",
		MediaType:   "media_doc",
	}

	media, err := domain.NewMediaEntity(mediaParams)
	assert.Nil(t, err)
	assert.Nil(t, localRepo.Save(media.ToMediaAggregate()))

	media2, err := domain.NewMediaEntity(mediaParam2)
	assert.Nil(t, err)
	assert.Nil(t, localRepo.Save(media2.ToMediaAggregate()))

	// Request page_token 1 (should start at) and page_size should be 1
	// Must return 2 objects, the requested one and the next_page_token one
	pagParam := &util.PaginationParams{
		TokenID:   0,
		TokenUUID: media.ExternalID.Value,
		Size:      1,
	}
	medias, err := localRepo.Fetch(pagParam, nil) // Should return 1 object from page_token 1
	assert.Nil(t, err)
	assert.True(t, len(medias) > 1)
}

func TestMediaLocalRepository_FetchByID(t *testing.T) {
	// Store a media
	localRepo := NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), new(MockLogger))
	media, err := domain.NewMediaEntity(mediaParams)
	assert.Nil(t, err)
	assert.Nil(t, localRepo.Save(media.ToMediaAggregate()))

	mediaSearched, err := localRepo.FetchByID(0, media.ExternalID.Value) // Should return previous stored media
	assert.Nil(t, err)
	assert.Equal(t, media.ExternalID.Value, mediaSearched.ExternalID)

	// Assert entity not found
	err = localRepo.RemoveOne(0, media.ExternalID.Value)
	assert.Nil(t, err)
	_, err = localRepo.FetchByID(0, media.ExternalID.Value)
	assert.True(t, errors.Is(err, global.EntityNotFound))
}

func TestMediaLocalRepository_FetchByTitle(t *testing.T) {
	// Store a media
	localRepo := NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), new(MockLogger))
	media, err := domain.NewMediaEntity(mediaParams)
	assert.Nil(t, err)
	assert.Nil(t, localRepo.Save(media.ToMediaAggregate()))

	mediaSearched, err := localRepo.FetchByTitle(media.Title.Value) // Should return previous stored media
	assert.Nil(t, err)
	assert.Equal(t, media.Title.Value, mediaSearched.Title)

	// Assert entity not found
	err = localRepo.RemoveOne(0, media.ExternalID.Value)
	assert.Nil(t, err)
	_, err = localRepo.FetchByTitle(media.Title.Value)
	assert.True(t, errors.Is(err, global.EntityNotFound))
}

func TestMediaLocalRepository_RemoveOne(t *testing.T) {
	// Store a media
	localRepo := NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), new(MockLogger))
	media, err := domain.NewMediaEntity(mediaParams)
	assert.Nil(t, err)
	assert.Nil(t, localRepo.Save(media.ToMediaAggregate()))

	err = localRepo.RemoveOne(0, media.ExternalID.Value)
	assert.Nil(t, err)
	_, err = localRepo.Fetch(nil, nil)
	assert.True(t, errors.Is(err, global.EntitiesNotFound))
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

package infrastructure

import (
	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/infrastructure/logging"
	"testing"
)

func TestCreateMediaLocal(t *testing.T) {
	logger, cleanup, _ := logging.NewLogger()
	defer cleanup()

	params := &domain.MediaEntityParams{
		Title:       "The foo programming language",
		DisplayName: "Foo book",
		Description: "The foo programming language has become one of the top notch...",
		UserID:      "60f90323-fc78-45e4-a0f5-71b63dd87d1a",
		AuthorID:    "a38d10fa-f369-4e8c-8c9d-f7f9f22bdc71",
		PublishDate: "2006-12-31",
		MediaType:   "media_book",
	}
	media, err := domain.NewMediaEntity(params)
	if err != nil {
		t.Error(err)
	}

	repository := NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), logger)
	err = repository.Save(media.ToMediaAggregate())
	if err != nil {
		t.Error(err)
	}
}

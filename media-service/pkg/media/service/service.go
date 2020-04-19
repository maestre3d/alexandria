package service

import (
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
)

type IMediaService interface {
	Create(title, displayName, description, userID, authorID, publishDate, mediaType string) (*domain.MediaEntity, error)
	List(pageToken, pageSize string, filterParams util.FilterParams) ([]*domain.MediaEntity, string, error)
	Get(id string) (*domain.MediaEntity, error)
	Update(id, title, displayName, description, userID, authorID, publishDate, mediaType string) (*domain.MediaEntity, error)
	Delete(id string) error
}

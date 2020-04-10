package application

import (
	"errors"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
)

type MediaUseCase struct {
	logger     util.ILogger
	repository domain.IMediaRepository
}

// MediaParams Required parameters to create/update a media
type MediaParams struct {
	MediaID     string
	Title       string
	DisplayName string
	Description string
	UserID      string
	AuthorID    string
	PublishDate string
	MediaType   string
}

func NewMediaUseCase(logger util.ILogger, repository domain.IMediaRepository) *MediaUseCase {
	return &MediaUseCase{logger, repository}
}

func (m *MediaUseCase) Create(params *MediaParams) error {
	mediaParams := &domain.MediaEntityParams{
		Title:       params.Title,
		DisplayName: params.DisplayName,
		Description: params.Description,
		UserID:      params.UserID,
		AuthorID:    params.AuthorID,
		PublishDate: params.PublishDate,
		MediaType:   params.MediaType,
	}
	media, err := domain.NewMediaEntity(mediaParams)
	if err != nil {
		return err
	}

	// Check media's title uniqueness
	existingMedia, err := m.GetByTitle(media.Title.Value)
	if errors.Is(err, global.EntityNotFound) {
		return err
	} else if existingMedia != nil {
		return global.EntityExists
	}

	return m.repository.Save(media.ToMediaAggregate())
}

func (m *MediaUseCase) GetByID(idString string) (*domain.MediaAggregate, error) {
	id, err := util.SanitizeID(idString)
	if err != nil {
		id = 0
	}
	err = util.SanitizeUUID(idString)
	if err != nil {
		if id <= 0 {
			return nil, err
		}
		idString = ""
	}

	return m.repository.FetchByID(id, idString)
}

func (m *MediaUseCase) GetByTitle(title string) (*domain.MediaAggregate, error) {
	if title == "" {
		return nil, global.EmptyQuery
	}

	return m.repository.FetchByTitle(title)
}

func (m *MediaUseCase) GetAll(params *util.PaginationParams) ([]*domain.MediaAggregate, error) {
	return m.repository.Fetch(params)
}

func (m *MediaUseCase) UpdateOneAtomic(params *MediaParams) error {
	if params == nil {
		return global.EmptyBody
	}

	id, err := util.SanitizeID(params.MediaID)
	if err != nil {
		id = 0
	}
	err = util.SanitizeUUID(params.MediaID)
	if err != nil {
		if id <= 0 {
			return err
		}
		params.MediaID = ""
	}

	mediaParams := &domain.MediaEntityParams{
		Title:       params.Title,
		DisplayName: params.DisplayName,
		Description: params.Description,
		UserID:      params.UserID,
		AuthorID:    params.AuthorID,
		PublishDate: params.PublishDate,
		MediaType:   params.MediaType,
	}
	media, err := domain.NewMediaEntity(mediaParams)
	if err != nil {
		return err
	}

	return m.repository.UpdateOne(id, params.MediaID, media.ToMediaAggregate())
}

func (m *MediaUseCase) UpdateOne(params *MediaParams) error {
	if params == nil {
		return global.EmptyBody
	}

	id, err := util.SanitizeID(params.MediaID)
	if err != nil {
		id = 0
	}
	err = util.SanitizeUUID(params.MediaID)
	if err != nil {
		if id <= 0 {
			return err
		}
		params.MediaID = ""
	}

	media, err := m.repository.FetchByID(id, params.MediaID)
	if err != nil {
		return err
	}

	// Ensure PATCH RPC/HTTP verb
	switch {
	case params.Title != "":
		media.Title = params.Title
	case params.MediaType != "":
		media.MediaType = params.MediaType
	case params.DisplayName != "":
		media.DisplayName = params.DisplayName
	case params.Description != "":
		media.Description = &params.Description
	case params.PublishDate != "":
		publishDate, err := domain.ParsePublishDate(params.PublishDate)
		if err != nil {
			return err
		}
		media.PublishDate = publishDate
	case params.UserID != "":
		media.UserID = params.UserID
	case params.AuthorID != "":
		media.AuthorID = params.AuthorID
	}

	err = media.ToMediaEntity().IsValid()
	if err != nil {
		return err
	}

	return m.repository.UpdateOne(id, params.MediaID, media)
}

func (m *MediaUseCase) RemoveOne(idString string) error {
	id, err := util.SanitizeID(idString)
	if err != nil {
		id = 0
	}
	err = util.SanitizeUUID(idString)
	if err != nil {
		if id <= 0 {
			return err
		}
		idString = ""
	}

	return m.repository.RemoveOne(id, idString)
}

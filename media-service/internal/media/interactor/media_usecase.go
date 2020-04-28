package interactor

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/global"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
)

type MediaUseCase struct {
	logger     log.Logger
	repository domain.IMediaRepository
}

func NewMediaUseCase(logger log.Logger, repository domain.IMediaRepository) *MediaUseCase {
	return &MediaUseCase{logger, repository}
}

func (u *MediaUseCase) Create(title, displayName, description, userID, authorID, publishDate, mediaType string) (*domain.MediaEntity, error) {
	// Validate
	var descriptionP *string
	descriptionP = nil
	if description != "" {
		descriptionP = &description
	}
	publish, err := time.Parse(global.RFC3339Micro, publishDate)
	if err != nil {
		return nil, fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"publish_date", global.RFC3339Micro))
	}

	mediaParams := &domain.MediaEntityParams{
		Title:       title,
		DisplayName: displayName,
		Description: descriptionP,
		UserID:      userID,
		AuthorID:    authorID,
		PublishDate: publish,
		MediaType:   mediaType,
	}
	media := domain.NewMediaEntity(mediaParams)
	err = media.IsValid()
	if err != nil {
		return nil, err
	}

	// Check title uniqueness
	existingMedia, _, err := u.List("0", "1", util.FilterParams{"title": title})
	if err == nil && len(existingMedia) > 0 {
		return nil, exception.EntityExists
	}

	err = u.repository.Save(media)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.media.created" domain event

	return media, nil
}

func (u *MediaUseCase) List(pageToken, pageSize string, filterParams util.FilterParams) (output []*domain.MediaEntity, nextToken string, err error) {
	params := util.NewPaginationParams(pageToken, pageSize)
	output, err = u.repository.Fetch(params, filterParams)

	nextToken = ""
	if len(output) >= params.Size {
		nextToken = output[len(output)-1].ExternalID
		output = output[0 : len(output)-1]
	}
	return
}

func (u *MediaUseCase) Get(id string) (*domain.MediaEntity, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, exception.InvalidID
	}

	return u.repository.FetchByID(id)
}

func (u *MediaUseCase) Update(id, title, displayName, description, userID, authorID, publishDate, mediaType string) (*domain.MediaEntity, error) {
	if id == "" && title == "" && displayName == "" && userID == "" && authorID == "" && publishDate == "" && mediaType == "" {
		return nil, exception.EmptyBody
	}

	// Get previous version
	media, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	// Update entity dynamically
	switch {
	case title != "":
		existingMedia, _, err := u.List("0", "1", util.FilterParams{"title": title})
		if err == nil && len(existingMedia) > 0 {
			return nil, exception.EntityExists
		}
		media.Title = title
	case mediaType != "":
		media.MediaType = strings.ToUpper(mediaType)
	case displayName != "":
		media.DisplayName = displayName
	case description != "":
		media.Description = &description
	case publishDate != "":
		publish, err := time.Parse(global.RFC3339Micro, publishDate)
		if err != nil {
			return nil, fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"publish_date", global.RFC3339Micro))
		}
		media.PublishDate = publish
	case userID != "":
		media.UserID = userID
	case authorID != "":
		media.AuthorID = authorID
	}

	err = media.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Update(media)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.media.updated" domain event

	return media, nil
}

func (u *MediaUseCase) Delete(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidID
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.media.deleted" domain event

	return u.repository.Remove(id)
}

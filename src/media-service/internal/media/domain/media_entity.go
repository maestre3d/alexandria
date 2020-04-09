package domain

import (
	"fmt"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"go.uber.org/multierr"
	"strings"
	"time"
)

// MediaEntity Media entity model
type MediaEntity struct {
	MediaID     mediaID
	ExternalID  externalID
	Title       title
	DisplayName displayName
	Description *description
	UserID      userID
	AuthorID    authorID
	PublishDate time.Time
	MediaType   mediaType
	CreateTime  time.Time
	UpdateTime  time.Time
	DeleteTime  *time.Time
	Metadata    *string
	Deleted     bool
}

// MediaEntityParams Required parameters to create an entity
type MediaEntityParams struct {
	Title       string
	DisplayName string
	Description string
	UserID      string
	AuthorID    string
	PublishDate string
	MediaType   string
}

func NewMediaEntity(params *MediaEntityParams) (*MediaEntity, error) {

	publishTime, err := time.Parse(global.RFC3339Micro, params.PublishDate)
	if err != nil {
		return nil, fmt.Errorf("%w:%s", global.InvalidFieldFormat, fmt.Sprintf(global.InvalidFieldFormatString, "publish_date", "date format 2006-01-02"))
	}

	descriptionPointer := &params.Description
	if params.Description == "" {
		descriptionPointer = nil
	}

	params.MediaType = strings.ToUpper(params.MediaType)

	media := &MediaEntity{
		MediaID:     mediaID{},
		ExternalID:  externalID{},
		Title:       title{Value: params.Title},
		DisplayName: displayName{Value: params.DisplayName},
		Description: &description{Value: descriptionPointer},
		UserID:      userID{Value: params.UserID},
		AuthorID:    authorID{Value: params.AuthorID},
		PublishDate: publishTime,
		MediaType:   mediaType{Value: params.MediaType},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		DeleteTime:  nil,
		Metadata:    nil,
		Deleted:     false,
	}
	media.ExternalID.Generate()

	err = media.IsValid()
	if err != nil {
		return nil, err
	}

	return media, nil
}

func (m *MediaEntity) IsValid() error {
	return multierr.Combine(
		m.ExternalID.IsValid(),
		m.Title.IsValid(),
		m.DisplayName.IsValid(),
		m.Description.IsValid(),
		m.UserID.IsValid(),
		m.AuthorID.IsValid(),
		m.MediaType.IsValid(),
	)
}

func (m *MediaEntity) ToMediaAggregate() *MediaAggregate {
	return &MediaAggregate{
		MediaID:     m.MediaID.Value,
		ExternalID:  m.ExternalID.Value,
		Title:       m.Title.Value,
		DisplayName: m.DisplayName.Value,
		Description: m.Description.Value,
		UserID:      m.UserID.Value,
		AuthorID:    m.AuthorID.Value,
		PublishDate: m.PublishDate,
		MediaType:   m.MediaType.Value,
		CreateTime:  m.CreateTime,
		UpdateTime:  m.UpdateTime,
		DeleteTime:  m.DeleteTime,
		Metadata:    m.Metadata,
		Deleted:     m.Deleted,
	}
}

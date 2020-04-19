package domain

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/exception"
	"strings"
	"time"
)

type mediaTypeEnum string

const (
	book    mediaTypeEnum = "MEDIA_BOOK"
	doc                   = "MEDIA_DOC"
	podcast               = "MEDIA_PODCAST"
	video                 = "MEDIA_VIDEO"
)

// MediaEntity Media entity type
type MediaEntity struct {
	MediaID     int64      `json:"-"`
	ExternalID  string     `json:"media_id"`
	Title       string     `json:"title"`
	DisplayName string     `json:"display_name"`
	Description *string    `json:"description"`
	UserID      string     `json:"user_id"`
	AuthorID    string     `json:"author_id"`
	PublishDate time.Time  `json:"publish_date"`
	MediaType   string     `json:"media_type"`
	CreateTime  time.Time  `json:"create_time"`
	UpdateTime  time.Time  `json:"update_time"`
	DeleteTime  *time.Time `json:"-"`
	Metadata    *string    `json:"metadata,omitempty"`
	Deleted     bool       `json:"-"`
}

type MediaEntityParams struct {
	Title       string
	DisplayName string
	Description *string
	UserID      string
	AuthorID    string
	PublishDate time.Time
	MediaType   string
}

func NewMediaEntity(params *MediaEntityParams) *MediaEntity {
	if params == nil {
		return nil
	}

	return &MediaEntity{
		MediaID:     0,
		ExternalID:  uuid.New().String(),
		Title:       params.Title,
		DisplayName: params.DisplayName,
		Description: params.Description,
		UserID:      params.UserID,
		AuthorID:    params.AuthorID,
		PublishDate: params.PublishDate,
		MediaType:   strings.ToUpper(params.MediaType),
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		DeleteTime:  nil,
		Metadata:    nil,
		Deleted:     false,
	}
}

func (e *MediaEntity) IsValid() error {
	if len(e.ExternalID) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "external_id"))
	}
	_, err := uuid.Parse(e.ExternalID)
	if err != nil {
		return fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.RequiredFieldString, "external_id", "uuid"))
	}

	if len(e.Title) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "title"))
	} else if len(e.Title) > 255 {
		return fmt.Errorf("%w:%s", exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "title", "1", "255"))
	}

	if len(e.DisplayName) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "display_name"))
	} else if len(e.DisplayName) > 100 {
		return fmt.Errorf("%w:%s", exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "display_name", "1", "100"))
	}

	if e.Description != nil && (len(*e.Description) == 0 || len(*e.Description) > 1024) {
		return fmt.Errorf("%w:%s", exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "description", "1", "1024"))
	}

	if len(e.UserID) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "user_id"))
	}
	_, err = uuid.Parse(e.UserID)
	if err != nil {
		return fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString, "user_id", "uuid"))
	}

	if len(e.AuthorID) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "author_id"))
	}
	_, err = uuid.Parse(e.AuthorID)
	if err != nil {
		return fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString, "author_id", "uuid"))
	}

	if len(e.MediaType) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "media_type"))
	} else if e.MediaType != string(book) && e.MediaType != string(doc) && e.MediaType != string(podcast) && e.MediaType != string(video) {
		return fmt.Errorf("%w:%s", exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString, "media_type", "[MEDIA_BOOK, MEDIA_DOC, MEDIA_PODCAST, MEDIA_VIDEO)"))
	}

	return nil
}

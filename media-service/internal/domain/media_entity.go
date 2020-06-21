package domain

import (
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-playground/validator/v10"
	gonanoid "github.com/matoous/go-nanoid"
	"strings"
	"time"
)

const (
	StatusDone    = "STATUS_DONE"
	StatusPending = "STATUS_PENDING"
	Book          = "MEDIA_BOOK"
	Podcast       = "MEDIA_PODCAST"
	Doc           = "MEDIA_DOC"
	Video         = "MEDIA_VIDEO"
)

type Media struct {
	ID          int64  `json:"-"`
	ExternalID  string `json:"id" validate:"required"`
	Title       string `json:"title" validate:"required,min=1,max=255"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=128"`
	Description string `json:"description" validate:"max=1024"`
	// ISO 639-2 Alpha-3 lang code
	LanguageCode string `json:"language_code" validate:"alphaunicode"`
	// User ID, Required SAGA transaction
	PublisherID string `json:"publisher_id"`
	// Required SAGA transaction
	AuthorID string `json:"author_id"`
	// Content's publish date
	PublishDate time.Time  `json:"publish_date"`
	MediaType   string     `json:"media_type" validate:"required,oneof=MEDIA_BOOK MEDIA_VIDEO MEDIA_DOC MEDIA_PODCAST"`
	CreateTime  time.Time  `json:"create_time"`
	UpdateTime  time.Time  `json:"update_time"`
	DeleteTime  *time.Time `json:"delete_time"`
	Active      bool       `json:"active"`
	ContentURL  *string    `json:"content_url"`
	TotalViews  int64      `json:"total_views"`
	Status      string     `json:"status" validate:"required,oneof=STATUS_DONE STATUS_PENDING"`
}

func NewMedia(ag *MediaAggregate) (*Media, error) {
	id, err := gonanoid.ID(16)
	if err != nil {
		return nil, err
	}

	if ag.DisplayName == "" {
		ag.DisplayName = ag.Title
	}

	ag.MediaType = ParseMediaType(ag.MediaType)
	publishDate, err := ParseDate(ag.PublishDate)
	if err != nil {
		return nil, err
	}

	return &Media{
		ID:           0,
		ExternalID:   id,
		Title:        ag.Title,
		DisplayName:  ag.DisplayName,
		Description:  ag.Description,
		LanguageCode: strings.ToLower(ag.LanguageCode),
		PublisherID:  ag.PublisherID,
		AuthorID:     ag.AuthorID,
		PublishDate:  publishDate,
		MediaType:    ag.MediaType,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
		DeleteTime:   nil,
		Active:       true,
		ContentURL:   nil,
		TotalViews:   0,
		Status:       StatusPending,
	}, nil
}

func ParseDate(date string) (time.Time, error) {
	publishDate, err := time.Parse(core.RFC3339Micro, date)
	if err != nil {
		return time.Time{}, exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"publish_date", "date format 2006-01-02"))
	}

	return publishDate, nil
}

func ParseMediaType(media string) string {
	media = strings.ToUpper(media)
	if media != Book && media != Podcast && media != Doc && media != Video {
		media = strings.ToLower(media)
		switch media {
		case "book":
			media = Book
		case "doc":
			media = Doc
		case "video":
			media = Video
		case "podcast":
			media = Podcast
		default:
			return ""
		}
	}

	return media
}

func (m *Media) IsValid() error {
	// Struct validation
	validate := validator.New()

	err := validate.Struct(m)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch {
			case err.Tag() == "required":
				return exception.NewErrorDescription(exception.RequiredField,
					fmt.Sprintf(exception.RequiredFieldString, strings.ToLower(err.Field())))
			case err.Tag() == "alphanum" || err.Tag() == "alpha" || err.Tag() == "alphanumunicode" || err.Tag() == "alphaunicode":
				return exception.NewErrorDescription(exception.InvalidFieldFormat,
					fmt.Sprintf(exception.InvalidFieldFormatString, strings.ToLower(err.Field()), err.Tag()))
			case err.Tag() == "max" || err.Tag() == "min":
				field := strings.ToLower(err.Field())
				maxLength := "n"

				switch field {
				case "title":
					maxLength = "255"
					break
				case "displayname":
					maxLength = "128"
					break
				case "description":
					maxLength = "1024"
					break
				}

				return exception.NewErrorDescription(exception.InvalidFieldRange,
					fmt.Sprintf(exception.InvalidFieldRangeString, field, "1", maxLength))
			case err.Tag() == "oneof":
				return exception.NewErrorDescription(exception.InvalidFieldFormat,
					fmt.Sprintf(exception.InvalidFieldFormatString, strings.ToLower(err.Field()),
						"["+err.Param()+"]"))
			}
		}

		return err
	}

	return nil
}

package domain

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
)

type mediaTypeEnum string

const (
	book    mediaTypeEnum = "MEDIA_BOOK"
	doc                   = "MEDIA_DOC"
	podcast               = "MEDIA_PODCAST"
	video                 = "MEDIA_VIDEO"
)

// MediaID Internal Unique Identifier
type mediaID struct {
	Value int64
}

func (m *mediaID) IsValid() error {
	if m.Value <= 0 {
		return fmt.Errorf("%w:%s", global.InvalidID, "media_id")
	}

	return nil
}

// ExternalID Unique Identifier for external services
type externalID struct {
	Value string
}

func (e *externalID) Generate() {
	e.Value = uuid.New().String()
}

func (e *externalID) IsValid() error {
	if e.Value == "" {
		return fmt.Errorf("%w:%s", global.RequiredField, fmt.Sprintf(global.RequiredFieldString, "external_id"))
	}

	_, err := uuid.Parse(e.Value)
	if err != nil {
		return fmt.Errorf("%w:%s", global.InvalidFieldFormat, fmt.Sprintf(global.RequiredFieldString, "external_id", "uuid"))
	}

	return nil
}

// Title Resource formal name
type title struct {
	Value string
}

func (t *title) IsValid() error {
	if t.Value == "" {
		return fmt.Errorf("%w:%s", global.RequiredField, fmt.Sprintf(global.RequiredFieldString, "title"))
	} else if len(t.Value) == 0 || len(t.Value) >= 256 {
		return fmt.Errorf("%w:%s", global.InvalidFieldRange, fmt.Sprintf(global.InvalidFieldRangeString, "title", "1", "255"))
	}
	return nil
}

// DisplayName Resource informal name
type displayName struct {
	Value string
}

func (t *displayName) IsValid() error {
	if t.Value == "" {
		return fmt.Errorf("%w:%s", global.RequiredField, fmt.Sprintf(global.RequiredFieldString, "display_name"))
	} else if len(t.Value) == 0 || len(t.Value) > 100 {
		return fmt.Errorf("%w:%s", global.InvalidFieldRange, fmt.Sprintf(global.InvalidFieldRangeString, "display_name", "1", "100"))
	}

	return nil
}

// Description Media description
type description struct {
	Value *string
}

func (d *description) IsValid() error {
	if d.Value != nil && (len(*d.Value) == 0 || len(*d.Value) > 1024) {
		return fmt.Errorf("%w:%s", global.InvalidFieldRange, fmt.Sprintf(global.InvalidFieldRangeString, "description", "1", "1024"))
	}

	return nil
}

// UserID User External Unique Identifier
type userID struct {
	Value string
}

func (u *userID) IsValid() error {
	if u.Value == "" {
		return fmt.Errorf("%w:%s", global.RequiredField, fmt.Sprintf(global.RequiredFieldString, "user_id"))
	}

	_, err := uuid.Parse(u.Value)
	if err != nil {
		return fmt.Errorf("%w:%s", global.InvalidFieldFormat, fmt.Sprintf(global.InvalidFieldFormatString, "user_id", "uuid"))
	}

	return nil
}

// AuthorID Author External Unique Identifier
type authorID struct {
	Value string
}

func (a *authorID) IsValid() error {
	if a.Value == "" {
		return fmt.Errorf("%w:%s", global.RequiredField, fmt.Sprintf(global.RequiredFieldString, "author_id"))
	}

	_, err := uuid.Parse(a.Value)
	if err != nil {
		return fmt.Errorf("%w:%s", global.InvalidFieldFormat, fmt.Sprintf(global.InvalidFieldFormatString, "author_id", "uuid"))
	}

	return nil
}

// MediaType Media resource type
type mediaType struct {
	Value string
}

func (m *mediaType) IsValid() error {
	if m.Value == "" {
		return fmt.Errorf("%w:%s", global.RequiredField, fmt.Sprintf(global.RequiredFieldString, "media_type"))
	} else if m.Value != string(book) && m.Value != string(doc) && m.Value != string(podcast) && m.Value != string(video) {
		return fmt.Errorf("%w:%s", global.InvalidFieldFormat, fmt.Sprintf(global.InvalidFieldFormatString, "media_type", "[MEDIA_BOOK, MEDIA_DOC, MEDIA_PODCAST, MEDIA_VIDEO)"))
	}

	return nil
}

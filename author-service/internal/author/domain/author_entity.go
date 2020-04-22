package domain

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"time"
)

// AuthorEntity Media's author
type AuthorEntity struct {
	AuthorID    int64      `json:"-"`
	ExternalID  string     `json:"author_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	DisplayName string     `json:"display_name"`
	BirthDate   time.Time  `json:"birth_date"`
	CreateTime  time.Time  `json:"create_time"`
	UpdateTime  time.Time  `json:"update_time"`
	DeleteTime  *time.Time `json:"-"`
	Metadata    *string    `json:"metadata,omitempty"`
	Deleted     bool       `json:"-"`
}

// NewAuthorEntity Create a new author
func NewAuthorEntity(firstName, lastName, displayName string, birth time.Time) *AuthorEntity {
	if displayName == "" {
		displayName = firstName + " " + lastName
	}

	return &AuthorEntity{
		AuthorID:    0,
		ExternalID:  uuid.New().String(),
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		BirthDate:   birth,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		DeleteTime:  nil,
		Metadata:    nil,
		Deleted:     false,
	}
}

func (e *AuthorEntity) IsValid() error {
	if len(e.FirstName) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "first_name"))
	} else if len(e.FirstName) > 255 {
		return fmt.Errorf("%w:%s", exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "first_name", "1", "255"))
	}

	if len(e.LastName) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "last_name"))
	} else if len(e.LastName) > 255 {
		return fmt.Errorf("%w:%s", exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "last_name", "1", "255"))
	}

	if len(e.DisplayName) == 0 {
		return fmt.Errorf("%w:%s", exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "display_name"))
	} else if len(e.DisplayName) > 255 {
		return fmt.Errorf("%w:%s", exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "display_name", "1", "255"))
	}

	return nil
}

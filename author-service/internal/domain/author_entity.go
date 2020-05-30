package domain

import (
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	gonanoid "github.com/matoous/go-nanoid"
	"time"
)

// Author Media's author
type Author struct {
	AuthorID    uint64     `json:"-"`
	ExternalID  string     `json:"author_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	DisplayName string     `json:"display_name"`
	BirthDate   time.Time  `json:"birth_date"`
	OwnerID     *string    `json:"owner_id"`
	CreateTime  time.Time  `json:"create_time"`
	UpdateTime  time.Time  `json:"update_time"`
	DeleteTime  *time.Time `json:"-"`
	Status      string     `json:"status"`
	Metadata    *string    `json:"metadata,omitempty"`
	Deleted     bool       `json:"-"`
}

// NewAuthor Create a new author
func NewAuthor(firstName, lastName, displayName, owner string, birth time.Time) *Author {
	if displayName == "" {
		displayName = firstName + " " + lastName
	}

	id, err := gonanoid.ID(16)
	if err != nil {
		return nil
	}

	return &Author{
		AuthorID:    core.NewSonyflakeID(),
		ExternalID:  id,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		BirthDate:   birth,
		OwnerID:     &owner,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		DeleteTime:  nil,
		Status:      StatusPending,
		Metadata:    nil,
		Deleted:     false,
	}
}

func (e *Author) IsValid() error {
	if len(e.FirstName) == 0 {
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "first_name"))
	} else if len(e.FirstName) > 255 {
		return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "first_name", "1", "255"))
	}

	if len(e.LastName) == 0 {
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "last_name"))
	} else if len(e.LastName) > 255 {
		return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "last_name", "1", "255"))
	}

	if len(e.DisplayName) == 0 {
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "display_name"))
	} else if len(e.DisplayName) > 255 {
		return exception.NewErrorDescription(exception.InvalidFieldRange, fmt.Sprintf(exception.InvalidFieldRangeString, "display_name", "1", "255"))
	}

	return nil
}

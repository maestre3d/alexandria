package domain

import (
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-playground/validator/v10"
	gonanoid "github.com/matoous/go-nanoid"
	"strings"
	"time"
)

// OwnershipType Owner's Type enum
type ownershipType string

const (
	CommunityOwner ownershipType = "public"
	PrivateOwner   ownershipType = "private"
	StatusPending                = "STATUS_PENDING"
	StatusDone                   = "STATUS_DONE"
)

// Author entity
type Author struct {
	ID            int64      `json:"-"`
	ExternalID    string     `json:"author_id"`
	FirstName     string     `json:"first_name" validate:"required,min=1,max=255,alphanumunicode"`
	LastName      string     `json:"last_name" validate:"required,min=1,max=255,alphanumunicode"`
	DisplayName   string     `json:"display_name" validate:"required,min=1,max=255"`
	OwnerID       string     `json:"owner_id" validate:"required"`
	OwnershipType string     `json:"ownership_type" validate:"required,oneof=public private"`
	CreateTime    time.Time  `json:"create_time"`
	UpdateTime    time.Time  `json:"update_time"`
	DeleteTime    *time.Time `json:"delete_time"`
	Active        bool       `json:"-"`
	Verified      bool       `json:"verified"`
	Picture       *string    `json:"picture"`
	TotalViews    int64      `json:"total_views"`
	Country       string     `json:"country" validate:"required,min=1,max=5,alphaunicode"`
	Status        string     `json:"status,omitempty" validate:"required,oneof=STATUS_PENDING STATUS_DONE"`
}

// NewAuthor Create a new author
func NewAuthor(firstName, lastName, displayName, ownershipType, ownerID, countryCode string) *Author {
	if displayName == "" {
		if firstName == "" {
			displayName = lastName
		} else if lastName == "" {
			displayName = firstName
		} else {
			displayName = firstName + " " + lastName
		}
	}

	if ownershipType == "" {
		ownershipType = string(PrivateOwner)
	} else {
		strings.ToLower(ownershipType)
	}

	// Generate external id
	id, err := gonanoid.ID(16)
	if err != nil {
		return nil
	}
	var picture string

	return &Author{
		ID:            0,
		ExternalID:    id,
		FirstName:     firstName,
		LastName:      lastName,
		DisplayName:   displayName,
		OwnerID:       ownerID,
		OwnershipType: ownershipType,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		DeleteTime:    nil,
		Active:        true,
		Verified:      false,
		Picture:       &picture,
		TotalViews:    0,
		Country:       countryCode,
		Status:        StatusPending,
	}
}

func (e Author) IsValid() error {
	// Struct validation
	validate := validator.New()

	err := validate.Struct(e)
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
				if field == "firstname" || field == "lastname" || field == "displayname" {
					return exception.NewErrorDescription(exception.InvalidFieldRange,
						fmt.Sprintf(exception.InvalidFieldRangeString, field, "1", "255"))
				}

				return exception.NewErrorDescription(exception.InvalidFieldRange,
					fmt.Sprintf(exception.InvalidFieldRangeString, field, "1", "n"))
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

package domain

import (
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-playground/validator/v10"
	gonanoid "github.com/matoous/go-nanoid"
	"strings"
	"time"
)

// RoleType Owner's role type enum
type roleType string

// OwnershipType Owner's Type enum
type ownershipType string

const (
	OwnerRole      roleType      = "owner"
	AdminRole      roleType      = "admin"
	ContribRole    roleType      = "contrib"
	CommunityOwner ownershipType = "public"
	PrivateOwner   ownershipType = "private"
)

// Owner represents user with permissions from the author
type Owner struct {
	ID   string `json:"owner_id"`
	Role string `json:"role" validate:"required,alphaunicode,oneof=owner admin contrib"`
}

func NewOwner(id, role string) *Owner {
	// Set root owner by default
	if role == "" {
		role = string(OwnerRole)
	}

	return &Owner{
		ID:   id,
		Role: strings.ToLower(role),
	}
}

// Author entity
type Author struct {
	ID            int64      `json:"-"`
	ExternalID    string     `json:"author_id"`
	FirstName     string     `json:"first_name" validate:"required,min=1,max=255,alphanumunicode"`
	LastName      string     `json:"last_name" validate:"required,min=1,max=255,alphanumunicode"`
	DisplayName   string     `json:"display_name" validate:"required,min=1,max=255"`
	OwnershipType string     `json:"ownership_type" validate:"required,oneof=public private"`
	CreateTime    time.Time  `json:"create_time"`
	UpdateTime    time.Time  `json:"update_time"`
	DeleteTime    *time.Time `json:"-"`
	Active        bool       `json:"-"`
	Verified      bool       `json:"verified"`
	Picture       *string    `json:"picture"`
	Owners        []*Owner   `json:"owners" validate:"required,min=1,dive"`
}

// NewAuthor Create a new author
func NewAuthor(firstName, lastName, displayName, ownershipType string, owner *Owner) *Author {
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

	owners := make([]*Owner, 0)

	// Add root owner
	if owner != nil {
		owner.Role = string(OwnerRole)
		owners = append(owners, owner)
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
		OwnershipType: ownershipType,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		DeleteTime:    nil,
		Active:        false,
		Verified:      false,
		Picture:       &picture,
		Owners:        owners,
	}
}

func (e *Author) IsValid() error {
	// Struct validation
	validate := validator.New()

	err := validate.Struct(e)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch {
			case err.Tag() == "required":
				return exception.NewErrorDescription(exception.RequiredField,
					fmt.Sprintf(exception.RequiredFieldString, strings.ToLower(err.Field())))
			case err.Tag() == "alphanum" || err.Tag() == "alpha" || err.Tag() == "alphanumunicode":
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

	// Compile-time owner's role enum assertion
	hasOwnerType := false
	for _, owner := range e.Owners {
		if owner.Role == string(OwnerRole) {
			hasOwnerType = true
			break
		}
	}

	// If author's user list doesn't has an owner, then entity's invalid
	if !hasOwnerType {
		return exception.NewErrorDescription(exception.RequiredField,
			fmt.Sprintf(exception.RequiredFieldString, "owner"))
	}

	return nil
}

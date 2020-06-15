package domain

import (
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-playground/validator/v10"
	"strings"
)

// RoleType Owner's role type enum
type roleType string

const (
	RootRole    roleType = "root"
	AdminRole   roleType = "admin"
	ContribRole roleType = "contrib"
)

// Owner represents user with permissions from the author
type Owner struct {
	ID   string `json:"owner_id" validate:"required"`
	Role string `json:"role" validate:"required,alphaunicode,oneof=root admin contrib"`
}

func NewOwner(id, role string) *Owner {
	// Set root owner by default
	if role == "" {
		role = string(RootRole)
	}

	return &Owner{
		ID:   id,
		Role: strings.ToLower(role),
	}
}

func (o Owner) IsValid() error {
	validate := validator.New()

	err := validate.Struct(o)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch {
			case err.Tag() == "required":
				return exception.NewErrorDescription(exception.RequiredField,
					fmt.Sprintf(exception.RequiredFieldString, strings.ToLower(err.Field())))
			case err.Tag() == "alphanum" || err.Tag() == "alpha" || err.Tag() == "alphanumunicode":
				return exception.NewErrorDescription(exception.InvalidFieldFormat,
					fmt.Sprintf(exception.InvalidFieldFormatString, strings.ToLower(err.Field()), err.Tag()))
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

package domain

import (
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-playground/validator/v10"
	gonanoid "github.com/matoous/go-nanoid"
	"strings"
	"time"
)

// Root/Default entity

type Category struct {
	ID         string
	ExternalID string    `json:"id"`
	Name       string    `json:"name" validate:"required,min=1,max=255"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	Active     bool      `json:"-"`
}

func NewCategory(name string) *Category {
	extID := ""
	id, err := gonanoid.Nanoid(16)
	if err == nil {
		extID = id
	}

	return &Category{
		ID:         "",
		ExternalID: extID,
		Name:       strings.Title(name),
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Active:     true,
	}
}

func (c Category) IsValid() error {
	validate := validator.New()

	err := validate.Struct(c)
	if err != nil {
		for _, errV := range err.(validator.ValidationErrors) {
			switch {
			case errV.Tag() == "required":
				return exception.NewErrorDescription(exception.RequiredField,
					fmt.Sprintf(exception.RequiredFieldString, strings.ToLower(errV.Field())))
			case errV.Tag() == "max" || errV.Tag() == "min":
				field := strings.ToLower(errV.Field())
				maxLength := "n"

				switch field {
				case "name":
					maxLength = "255"
					break
				}

				return exception.NewErrorDescription(exception.InvalidFieldRange,
					fmt.Sprintf(exception.InvalidFieldRangeString, field, "1", maxLength))
			}
		}

		return err
	}

	return nil
}

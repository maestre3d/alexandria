package domain

import (
	"fmt"
	"github.com/alexandria-oss/core/exception"
)

type Identity struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (i Identity) IsValid() error {
	switch {
	case i.Email == "":
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "email"))
	case i.Username == "":
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "username"))
	case i.Password == "":
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "password"))
	}

	return nil
}

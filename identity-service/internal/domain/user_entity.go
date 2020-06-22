package domain

import (
	"fmt"
	"github.com/alexandria-oss/core/exception"
)

type User struct {
	ID string `json:"id"`
	Email    string `json:"email,omitempty"`
	Name string `json:"name"`
	GivenName string `json:"given_name"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

func (i User) IsValid() error {
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


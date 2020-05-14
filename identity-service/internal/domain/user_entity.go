package domain

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type User struct {
	ID         int64     `json:"id"`
	ExternalID string    `json:"external_id"`
	Username   string    `json:"username" validate:"required,gte=4,lte=30,alphanum"`
	Name       string    `json:"name" validate:"required,gt=0,lte=50,alpha"`
	LastName   string    `json:"last_name" validate:"required,gt=0,lte=50,alpha"`
	Email      string    `json:"email" validate:"required,email"`
	Gender     string    `json:"gender" validate:"gt=0,lte=50,alpha,oneof=MALE FEMALE OTHER"`
	Locale     string    `json:"locale" validate:"required,gt=0,lte=7"`
	Picture    *string   `json:"picture" validate:"-"`
	Deleted    bool      `json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  time.Time `json:"deleted_at"`
}

func NewUser(username, name, lastName, email, gender, locale string) *User {
	picture := ""
	if gender != "" {
		gender = strings.ToUpper(gender)
	} else {
		gender = "MALE"
	}

	if locale != "" {
		locale = strings.ToUpper(locale)
	} else {
		locale = "EN-US"
	}

	user := &User{
		ID:         0,
		ExternalID: uuid.New().String(),
		Username:   strings.ToLower(username),
		Name:       name,
		LastName:   lastName,
		Email:      strings.ToLower(email),
		Gender:     gender,
		Locale:     locale,
		Picture:    &picture,
		Deleted:    false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		DeletedAt:  time.Time{},
	}

	user.NameSanitize()
	user.LastNameSanitize()

	return user
}

func (u *User) NameSanitize() {
	nameList := strings.Split(u.Name, "")
	nameList[0] = strings.ToUpper(nameList[0])

	u.Name = ""
	for _, item := range nameList {
		u.Name += item
	}
}

func (u *User) LastNameSanitize() {
	lastNameList := strings.Split(u.LastName, "")
	lastNameList[0] = strings.ToUpper(lastNameList[0])

	u.LastName = ""
	for _, item := range lastNameList {
		u.LastName += item
	}
}

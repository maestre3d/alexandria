package domain

import (
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/auth"
	"github.com/google/uuid"
	"strings"
	"time"
)

type User struct {
	// ID unique identifier using a distributed ID
	ID uint64 `json:"-"`
	// ExternalID unique identifier using base64 encoding or UUID
	ExternalID string     `json:"id"`
	Username   string     `json:"username" validate:"required,gte=4,lte=30,alphanum"`
	Password   string     `json:"password" validate:"required,gte=8,lte=64"`
	Name       string     `json:"name" validate:"required,gt=0,lte=50"`
	LastName   string     `json:"last_name" validate:"required,gt=0,lte=50"`
	Email      string     `json:"email" validate:"required,email"`
	Gender     string     `json:"gender" validate:"gt=0,lte=50,alpha,oneof=MALE FEMALE OTHER"`
	Locale     string     `json:"locale" validate:"required,gt=0,lte=7"`
	Picture    *string    `json:"picture"`
	Role       string     `json:"role" validate:"gt=0,lte=64,alpha,oneof=ROLE_ROOT ROLE_ADMIN ROLE_USER"`
	CreateTime time.Time  `json:"create_time"`
	UpdateTime time.Time  `json:"update_time"`
	DeleteTime *time.Time `json:"delete_time"`
	Deleted    bool       `json:"deleted"`
	Active     bool       `json:"active"`
	Verified   bool       `json:"verified"`
}

const (
	// GenderMale user's male gender
	GenderMale = "MALE"
	// GenderFemale user's female gender
	GenderFemale = "FEMALE"
	// GenderOther user's other gender
	GenderOther = "OTHER"
)

// NewUser returns a user entity with required and sanitized values
func NewUser(username, password, name, lastName, email, gender, locale, role string) *User {
	name = CapitalizeString(name)
	lastName = CapitalizeString(lastName)

	var picture string
	if gender != "" {
		gender = strings.ToUpper(gender)
	} else {
		gender = GenderMale
	}

	if locale != "" {
		locale = strings.ToUpper(locale)
	} else {
		locale = core.LocaleDefault
	}

	if role != "" {
		role = strings.ToUpper(role)
	} else {
		role = auth.RoleUser
	}

	// TODO: Change UUID for base64 encoded ID
	return &User{
		ID:         core.NewSonyflakeID(),
		ExternalID: uuid.New().String(),
		Username:   strings.ToLower(username),
		Password:   password,
		Name:       name,
		LastName:   lastName,
		Email:      strings.ToLower(email),
		Gender:     gender,
		Locale:     locale,
		Picture:    &picture,
		Role:       role,
		Deleted:    false,
		Active:     false,
		Verified:   false,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		DeleteTime: nil,
	}
}

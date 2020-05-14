package interactor

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/go-playground/validator/v10"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"strings"
	"time"
)

type UserUseCase struct {
	ctx        context.Context
	logger     log.Logger
	repository domain.UserRepository
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func NewUserUseCase(ctx context.Context, logger log.Logger, repository domain.UserRepository) *UserUseCase {
	return &UserUseCase{ctx, logger, repository}
}

func (u *UserUseCase) Create(username, name, lastName, email, gender, locale string) (*domain.User, error) {
	user := domain.NewUser(username, name, lastName, email, gender, locale)
	err := validate.Struct(user)
	if err != nil {
		return nil, err
	}

	err = u.repository.Save(*user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) Update(id, username, name, lastName, email, gender, locale string) (*domain.User, error) {
	user, err := u.GetByID(id)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, exception.EntityNotFound
	}

	// Update entity dynamically
	if username != "" {
		user.Username = username
	}
	if name != "" {
		user.Name = name
		user.NameSanitize()
	}
	if lastName != "" {
		user.LastName = lastName
		user.LastNameSanitize()
	}
	if email != "" {
		user.Email = email
	}
	if gender != "" {
		user.Gender = strings.ToUpper(gender)
	}
	if locale != "" {
		user.Locale = strings.ToUpper(locale)
	}
	user.UpdatedAt = time.Now()

	err = validate.Struct(user)
	if err != nil {
		return nil, err
	}

	err = u.repository.Replace(*user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) List() ([]*domain.User, string, error) {
	users, err := u.repository.Fetch(*core.NewPaginationParams("", ""), nil)
	if err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(users) > 1 {
		nextToken = users[len(users)-1].ExternalID
		users = users[0 : len(users)-1]
	}

	return users, nextToken, nil
}

func (u *UserUseCase) GetByID(id string) (*domain.User, error) {
	err := core.ValidateUUID(id)
	if err != nil {
		return nil, err
	}

	user, err := u.repository.FetchByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUseCase) Delete(id string) error {
	err := core.ValidateUUID(id)
	if err != nil {
		return err
	}

	return u.repository.Remove(id)
}

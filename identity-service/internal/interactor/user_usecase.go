package interactor

import (
	"context"
	"errors"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/crypto"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/go-playground/validator/v10"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"strconv"
	"strings"
	"time"
)

// UserUseCase contains business logic functions
type UserUseCase struct {
	logger     log.Logger
	repository domain.UserRepository
}

var validate *validator.Validate
var argon2config *crypto.Argon2Config

func init() {
	argon2config = crypto.DefaultArgon2Config()
}

// NewUserUseCase returns a user interactor
func NewUserUseCase(logger log.Logger, repository domain.UserRepository) *UserUseCase {
	validate = validator.New()
	return &UserUseCase{logger, repository}
}

// Create a new user entity
func (u *UserUseCase) Create(ctx context.Context, params domain.UserAggregate) (*domain.User, error) {
	user := domain.NewUser(params.Username, params.Password, params.Name, params.LastName, params.Email,
		params.Gender, params.Locale, params.Role)
	err := validate.Struct(user)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch {
			case err.Tag() == "required":
				return nil, exception.NewErrorDescription(exception.RequiredField,
					fmt.Sprintf(exception.RequiredFieldString, strings.ToLower(err.Field())))
			case err.Tag() == "alphanum" || err.Tag() == "alpha" || err.Tag() == "email":
				return nil, exception.NewErrorDescription(exception.InvalidFieldFormat,
					fmt.Sprintf(exception.InvalidFieldFormatString, strings.ToLower(err.Field()), err.Tag()))
			}
		}
		return nil, err
	}

	// Since our identity provider automatically encrypts our passwords,
	// we must hash our passwords right after we got our user inside our provider's pool
	user.Password = crypto.Argon2HashString(user.Password, argon2config)
	err = u.repository.Save(ctx, *user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update a user entity atomically
func (u *UserUseCase) Update(ctx context.Context, token string, params domain.UserAggregate, oldPassword string) (*domain.User, error) {
	user, err := u.Get(ctx, token)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, exception.EntityNotFound
	}

	// Update entity dynamically
	if params.Username != "" {
		user.Username = params.Username
	}
	if params.Password != "" {
		if oldPassword == "" {
			return nil, exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "old_password"))
		}

		if ok := crypto.Argon2CompareString(oldPassword, user.Password); !ok {
			return nil, errors.New("incorrect password")
		}
		user.Password = params.Password
	}
	if params.Name != "" {
		user.Name = domain.CapitalizeString(params.Name)
	}
	if params.LastName != "" {
		user.LastName = domain.CapitalizeString(params.LastName)
	}
	if params.Email != "" {
		user.Email = params.Email
	}
	if params.Gender != "" {
		user.Gender = strings.ToUpper(params.Gender)
	}
	if params.Locale != "" {
		user.Locale = strings.ToUpper(params.Locale)
	}
	if params.Role != "" {
		user.Role = strings.ToUpper(params.Role)
	}
	if params.Active != "" {
		state, err := strconv.ParseBool(params.Active)
		if err == nil {
			user.Active = state
		}
	}
	user.UpdateTime = time.Now()

	err = validate.Struct(user)
	if err != nil {
		return nil, err
	}

	// Since our identity provider automatically encrypts our passwords,
	// we must hash our passwords right after we got our user inside our provider's pool
	crypto.Argon2HashString(params.Password, argon2config)
	err = u.repository.Replace(ctx, *user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// List user's using pagination tokens and filtering
func (u *UserUseCase) List(ctx context.Context, params *core.PaginationParams, filter core.FilterParams) ([]*domain.User, string, error) {
	// Required to get next_page_token
	params.Size++

	users, err := u.repository.Fetch(ctx, *params, filter)
	if err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(users) >= params.Size {
		nextToken = users[len(users)-1].ExternalID
		users = users[0 : len(users)-1]
	}

	return users, nextToken, nil
}

// Get a single user
func (u *UserUseCase) Get(ctx context.Context, token string) (*domain.User, error) {
	return u.repository.FetchOne(ctx, token)
}

// Delete softly a user
func (u *UserUseCase) Delete(ctx context.Context, token string) error {
	return u.repository.Remove(ctx, token)
}

// ForceDelete remove a user hardly (hard delete)
func (u *UserUseCase) ForceDelete(ctx context.Context, token string) error {
	return u.repository.HardRemove(ctx, token)
}

package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
)

// IdentityUseCase identity provider's transactions
type IdentityUseCase struct {
	logger   log.Logger
	provider domain.ProviderAdapter
}

func NewIdentityUseCase(logger log.Logger, provider domain.ProviderAdapter) *IdentityUseCase {
	return &IdentityUseCase{
		logger:   logger,
		provider: provider,
	}
}

// SignUp register a new identity
func (u *IdentityUseCase) SignUp(ctx context.Context, params domain.UserAggregate) error {
	i := domain.Identity{
		Email:    params.Email,
		Username: params.Username,
		Password: params.Password,
	}

	// Validate identity, avoid remote fetching if not valid
	err := i.IsValid()
	if err != nil {
		return err
	}

	// Insert a registry inside our identity provider (i.e. AWS Cognito or Azure ActiveDirectory)
	// before any persistence transaction
	return u.provider.SignUp(ctx, i)
}

// ConfirmSignUp verify an identity sign up with OTP (Email/Phone verification code)
func (u *IdentityUseCase) ConfirmSignUp(ctx context.Context, username, code string) error {
	return u.provider.ConfirmSignUp(ctx, username, code)
	// TODO: Move this line to view/pkg
	// Once our identity provider gets notified and transaction
	// is successfully, then update our database
	// _, err = u.userCase.Update(ctx, username, domain.UserAggregate{Active: "true"}, "")
}

// SignIn get an access_token from identity's credentials/refresh_token
func (u *IdentityUseCase) SignIn(ctx context.Context, username, password, refreshToken, deviceKey string) (*domain.ProviderToken, error) {
	return u.provider.SignIn(ctx, username, password, refreshToken, deviceKey)
}

// SignOut removes identity's sessions from any device
func (u *IdentityUseCase) SignOut(ctx context.Context, accessToken string) error {
	return u.provider.SignOut(ctx, accessToken)
}

// Get return an identity entity from an access_token
func (u *IdentityUseCase) Get(ctx context.Context, accessToken string) (*domain.Identity, error) {
	// TODO: Parse to entity
	return u.provider.Get(ctx, accessToken)
}

// Update mutate an identity from an access_token
func (u *IdentityUseCase) Update(ctx context.Context, accessToken string, params domain.UserAggregate) error {
	i, err := u.Get(ctx, accessToken)
	if err != nil {
		return err
	}

	// Dynamic update
	if params.Email != "" {
		i.Email = params.Email
	}
	if params.Username != "" {
		i.Username = params.Username
	}
	if params.Password != "" {
		i.Password = params.Password
	}

	return u.provider.Update(ctx, accessToken, *i)
}

// Admin API

// ForceSignOut remove sessions from any device as Admin
func (u *IdentityUseCase) ForceSignOut(ctx context.Context, username string) error {
	return u.provider.ForceSignOut(ctx, username)
}

// ForceGet return an identity entity using a username as Admin
func (u *IdentityUseCase) ForceGet(ctx context.Context, username string) (*domain.Identity, error) {
	return u.provider.ForceGet(ctx, username)
}

// ForceUpdate mutate an identity using a username as Admin
func (u *IdentityUseCase) ForceUpdate(ctx context.Context, username string, params domain.UserAggregate) error {
	if username == "" {
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "username"))
	}

	i, err := u.ForceGet(ctx, params.Username)
	if err != nil {
		return err
	}

	// Dynamic update
	if params.Username != "" {
		i.Username = params.Username
	}
	if params.Password != "" {
		i.Password = params.Password
	}
	if params.Email != "" {
		i.Email = params.Email
	}

	return u.provider.ForceUpdate(ctx, *i)
}

// ForceDelete remove an identity as Admin
func (u *IdentityUseCase) ForceDelete(ctx context.Context, username string) error {
	return u.provider.ForceDelete(ctx, username)
}

func (u *IdentityUseCase) ForceEnable(ctx context.Context, username string) error {
	return u.provider.ForceEnable(ctx, username)
}

func (u *IdentityUseCase) ForceHardDelete(ctx context.Context, username string) error {
	return u.provider.ForceHardDelete(ctx, username)
}

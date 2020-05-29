package domain

import "context"

// ProviderToken token provided by the authentication process
type ProviderToken struct {
	ID           string `json:"token_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Type         string `json:"type"`
	DeviceKey    string `json:"device_key"`
}

// ProviderAdapter Identity 3rd-party provider
type ProviderAdapter interface {
	SignUp(ctx context.Context, identity Identity) error
	ConfirmSignUp(ctx context.Context, username, code string) error
	SignIn(ctx context.Context, username, password, refreshToken, deviceKey string) (*ProviderToken, error)
	SignOut(ctx context.Context, accessToken string) error
	Get(ctx context.Context, accessToken string) (*Identity, error)
	Update(ctx context.Context, accessToken string, identity Identity) error
	RegisterDevice(ctx context.Context, accessToken, deviceKey, deviceName string) error

	// Admin API
	ForceSignOut(ctx context.Context, username string) error
	ForceGet(ctx context.Context, username string) (*Identity, error)
	ForceUpdate(ctx context.Context, identity Identity) error
	ForceDelete(ctx context.Context, username string) error
	ForceEnable(ctx context.Context, username string) error
	ForceHardDelete(ctx context.Context, username string) error
}

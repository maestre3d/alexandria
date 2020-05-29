package infrastructure

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"sync"
)

// IdentityCognitoAdapter Identity provider's AWS Cognito implementation
type IdentityCognitoAdapter struct {
	client *cognito.CognitoIdentityProvider
	cfg    *config.Kernel
	logger log.Logger
	mtx    *sync.Mutex
}

func NewProviderCognitoAdapter(logger log.Logger, cfg *config.Kernel) *IdentityCognitoAdapter {
	return &IdentityCognitoAdapter{
		client: newCognitoClient(),
		cfg:    cfg,
		logger: logger,
		mtx:    new(sync.Mutex),
	}
}

func newCognitoClient() *cognito.CognitoIdentityProvider {
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return cognito.New(s)
}

func (a *IdentityCognitoAdapter) SignUp(ctx context.Context, identity domain.Identity) error {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	userData := &cognito.SignUpInput{
		Username: aws.String(identity.Username),
		Password: aws.String(identity.Password),
		ClientId: aws.String(a.cfg.AWS.CognitoClientID),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(identity.Email),
			},
		},
	}

	_, err := a.client.SignUpWithContext(ctx, userData)
	return err
}

func (a *IdentityCognitoAdapter) ConfirmSignUp(ctx context.Context, username, code string) error {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	userData := &cognito.ConfirmSignUpInput{
		ConfirmationCode: aws.String(code),
		Username:         aws.String(username),
		ClientId:         aws.String(a.cfg.AWS.CognitoClientID),
	}

	_, err := a.client.ConfirmSignUpWithContext(ctx, userData)
	return err
}

func (a *IdentityCognitoAdapter) SignIn(ctx context.Context, username, password, refreshToken, deviceKey string) (*domain.ProviderToken, error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Indicate which auth flow we want
	flowUsernamePassword := "USER_PASSWORD_AUTH"
	flowRefreshToken := "REFRESH_TOKEN_AUTH"

	flow := aws.String(flowUsernamePassword)
	var params map[string]*string

	if refreshToken != "" && deviceKey != "" {
		flow = aws.String(flowRefreshToken)
		params = map[string]*string{
			"REFRESH_TOKEN": aws.String(refreshToken),
			"DEVICE_KEY":    aws.String(deviceKey),
		}
	} else {
		params = map[string]*string{
			"USERNAME": aws.String(username),
			"PASSWORD": aws.String(password),
		}
	}

	authData := &cognito.InitiateAuthInput{
		AnalyticsMetadata: nil,
		AuthFlow:          flow,
		AuthParameters:    params,
		ClientId:          aws.String(a.cfg.AWS.CognitoClientID),
		ClientMetadata:    nil,
		UserContextData:   nil,
	}

	r, err := a.client.InitiateAuthWithContext(ctx, authData)
	if err != nil {
		return nil, err
	}

	// If MFA is activated, then opt to new challenge
	if r.Session != nil {
		challengeData := &cognito.RespondToAuthChallengeInput{
			ChallengeName:      r.ChallengeName,
			ChallengeResponses: r.ChallengeParameters,
			ClientId:           aws.String(a.cfg.AWS.CognitoClientID),
			Session:            r.Session,
		}

		rC, err := a.client.RespondToAuthChallengeWithContext(ctx, challengeData)
		if err != nil {
			return nil, err
		}

		if rC.AuthenticationResult != nil {
			token := &domain.ProviderToken{
				ID:          *r.AuthenticationResult.IdToken,
				AccessToken: *r.AuthenticationResult.AccessToken,
				ExpiresIn:   *r.AuthenticationResult.ExpiresIn,
				Type:        *r.AuthenticationResult.TokenType,
			}

			if rC.AuthenticationResult.RefreshToken != nil {
				token.RefreshToken = *rC.AuthenticationResult.RefreshToken
			}

			if rC.AuthenticationResult.NewDeviceMetadata != nil {
				token.DeviceKey = *r.AuthenticationResult.NewDeviceMetadata.DeviceKey
			}

			return token, nil
		}

		return nil, nil
	}

	// Avoid any error panicking during runtime
	if r.AuthenticationResult == nil {
		return nil, nil
	}

	token := &domain.ProviderToken{
		ID:          *r.AuthenticationResult.IdToken,
		AccessToken: *r.AuthenticationResult.AccessToken,
		ExpiresIn:   *r.AuthenticationResult.ExpiresIn,
		Type:        *r.AuthenticationResult.TokenType,
	}

	if r.AuthenticationResult.RefreshToken != nil {
		token.RefreshToken = *r.AuthenticationResult.RefreshToken
	}

	if r.AuthenticationResult.NewDeviceMetadata != nil {
		token.DeviceKey = *r.AuthenticationResult.NewDeviceMetadata.DeviceKey
	}

	return token, nil
}

func (a *IdentityCognitoAdapter) SignOut(ctx context.Context, accessToken string) error {
	signData := &cognito.GlobalSignOutInput{
		AccessToken: aws.String(accessToken),
	}

	_, err := a.client.GlobalSignOutWithContext(ctx, signData)
	return err
}

func (a *IdentityCognitoAdapter) Get(ctx context.Context, accessToken string) (*domain.Identity, error) {
	getData := &cognito.GetUserInput{
		AccessToken: aws.String(accessToken),
	}

	r, err := a.client.GetUserWithContext(ctx, getData)
	if err != nil {
		return nil, err
	}

	i := &domain.Identity{
		Username: *r.Username,
	}

	for _, atr := range r.UserAttributes {
		switch atr.Name {
		case aws.String("email"):
			i.Email = *atr.Value
			continue
		case aws.String("password"):
			i.Password = *atr.Value
			continue
		}
	}

	return i, nil
}

func (a *IdentityCognitoAdapter) Update(ctx context.Context, accessToken string, identity domain.Identity) error {
	updateData := &cognito.UpdateUserAttributesInput{
		AccessToken:    aws.String(accessToken),
		ClientMetadata: nil,
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(identity.Email),
			},
			{
				Name:  aws.String("password"),
				Value: aws.String(identity.Password),
			},
			{
				Name:  aws.String("username"),
				Value: aws.String(identity.Username),
			},
		},
	}

	_, err := a.client.UpdateUserAttributesWithContext(ctx, updateData)
	return err
}

func (a *IdentityCognitoAdapter) RegisterDevice(ctx context.Context, accessToken, deviceKey, deviceName string) error {
	deviceData := &cognito.ConfirmDeviceInput{
		AccessToken: aws.String(accessToken),
		DeviceKey:   aws.String(deviceKey),
		DeviceName:  aws.String(deviceName),
	}

	_, err := a.client.ConfirmDeviceWithContext(ctx, deviceData)
	if err != nil {
		return err
	}

	return nil
}

// Admin zone
// Forcing or volatile operations for Admin API

func (a *IdentityCognitoAdapter) ForceSignOut(ctx context.Context, username string) error {
	signData := &cognito.AdminUserGlobalSignOutInput{
		UserPoolId: aws.String(a.cfg.AWS.CognitoPoolID),
		Username:   aws.String(username),
	}

	_, err := a.client.AdminUserGlobalSignOutWithContext(ctx, signData)
	return err
}

func (a *IdentityCognitoAdapter) ForceGet(ctx context.Context, username string) (*domain.Identity, error) {
	getData := &cognito.AdminGetUserInput{
		UserPoolId: aws.String(a.cfg.AWS.CognitoPoolID),
		Username:   aws.String(username),
	}

	r, err := a.client.AdminGetUserWithContext(ctx, getData)
	if err != nil {
		return nil, err
	}

	i := &domain.Identity{
		Username: *r.Username,
	}

	for _, atr := range r.UserAttributes {
		switch atr.Name {
		case aws.String("email"):
			i.Email = *atr.Value
			break
		case aws.String("password"):
			i.Password = *atr.Value
			continue
		}
	}

	return i, nil
}

func (a *IdentityCognitoAdapter) ForceUpdate(ctx context.Context, identity domain.Identity) error {
	updateData := &cognito.AdminUpdateUserAttributesInput{
		ClientMetadata: nil,
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(identity.Email),
			},
			{
				Name:  aws.String("password"),
				Value: aws.String(identity.Password),
			},
			{
				Name:  aws.String("username"),
				Value: aws.String(identity.Username),
			},
		},
		UserPoolId: aws.String(a.cfg.AWS.CognitoPoolID),
		Username:   aws.String(identity.Username),
	}

	_, err := a.client.AdminUpdateUserAttributesWithContext(ctx, updateData)
	return err
}

func (a *IdentityCognitoAdapter) ForceDelete(ctx context.Context, username string) error {
	deleteData := &cognito.AdminDisableUserInput{
		Username:   aws.String(username),
		UserPoolId: aws.String(a.cfg.AWS.CognitoPoolID),
	}

	_, err := a.client.AdminDisableUserWithContext(ctx, deleteData)
	return err
}

func (a *IdentityCognitoAdapter) ForceEnable(ctx context.Context, username string) error {
	enableData := &cognito.AdminEnableUserInput{
		UserPoolId: aws.String(a.cfg.AWS.CognitoPoolID),
		Username:   aws.String(username),
	}

	_, err := a.client.AdminEnableUserWithContext(ctx, enableData)
	return err
}

func (a *IdentityCognitoAdapter) ForceHardDelete(ctx context.Context, username string) error {
	deleteData := &cognito.AdminDeleteUserInput{
		UserPoolId: aws.String(a.cfg.AWS.CognitoPoolID),
		Username:   aws.String(username),
	}

	_, err := a.client.AdminDeleteUserWithContext(ctx, deleteData)
	return err
}

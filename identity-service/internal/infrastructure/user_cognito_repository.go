package infrastructure

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/exception"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"sync"
)

// UserCognitoRepository User provider's AWS Cognito repo implementation
type UserCognitoRepository struct {
	client *cognito.CognitoIdentityProvider
	cfg    *config.Kernel
	logger log.Logger
	mu    *sync.Mutex
}

func NewUserCognitoRepository(logger log.Logger, cfg *config.Kernel) *UserCognitoRepository {
	return &UserCognitoRepository{
		client: newCognitoClient(),
		cfg:    cfg,
		logger: logger,
		mu:    new(sync.Mutex),
	}
}

func newCognitoClient() *cognito.CognitoIdentityProvider {
	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return cognito.New(s)
}

func (r *UserCognitoRepository) FetchByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	statement := fmt.Sprintf("sub = \"%s\"", id)

	i := &cognito.ListUsersInput{
		AttributesToGet: nil,
		Filter:          aws.String(statement),
		Limit:           aws.Int64(1),
		PaginationToken: nil,
		UserPoolId:      aws.String(r.cfg.AWS.CognitoPoolID),
	}

	userC, err := r.client.ListUsersWithContext(ctx, i)
	if err != nil {
		return nil, err
	}

	if len(userC.Users) == 0 {
		return nil, exception.EntityNotFound
	}

	user := &domain.User{
		Username: *userC.Users[0].Username,
	}

	for _, atr := range userC.Users[0].Attributes {
		switch *atr.Name {
		case "email":
			user.Email = *atr.Value
			continue
		case "name":
			user.Name = *atr.Value
			continue
		case "given_name":
			user.GivenName = *atr.Value
			continue
		case "sub":
			user.ID = *atr.Value
			continue
		}
	}

	return user, nil
}
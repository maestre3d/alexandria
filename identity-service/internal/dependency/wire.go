// +build wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/identity-service/internal/interactor"
)

var Ctx context.Context = context.Background()

var dataSet = wire.NewSet(
	provideContext,
	config.NewKernel,
	logger.NewZapLogger,
	wire.Bind(new(domain.UserRepository), new(*infrastructure.UserCognitoRepository)),
	infrastructure.NewUserCognitoRepository,
	// provide kafka/sqs base event bus
)

func provideContext() context.Context {
	return Ctx
}

func InjectUserUseCase() (*interactor.User, error) {
	wire.Build(
		dataSet,
		interactor.NewUser,
	)

	return &interactor.User{}, nil
}

func InjectUserSAGAUseCase() (*interactor.UserSAGA, error) {
	wire.Build(
		dataSet,
		wire.Bind(new(domain.UserEventSAGA), new(*infrastructure.UserSAGAKafkaEvent)),
		infrastructure.NewUserSAGAKafkaEvent,
		interactor.NewUserSAGA,
	)

	return &interactor.UserSAGA{}, nil
}

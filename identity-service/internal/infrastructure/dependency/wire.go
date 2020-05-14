//+build wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/identity-service/internal/interactor"
)

var persistenceSet = wire.NewSet(
	provideContext,
	config.NewKernelConfiguration,
	persistence.NewPostgresPool,
	persistence.NewRedisPool,
)

var userRepository = wire.NewSet(
	logger.NewZapLogger,
	persistenceSet,
	wire.Bind(new(domain.UserRepository), new(*infrastructure.UserPostgresRepository)),
	infrastructure.NewUserPostgresRepository,
)

var userUseCase = wire.NewSet(
	userRepository,
	interactor.NewUserUseCase,
)

func provideContext() context.Context {
	return context.Background()
}

func InjectUserUseCase() (*interactor.UserUseCase, func(), error) {
	wire.Build(userUseCase)
	return &interactor.UserUseCase{}, nil, nil
}

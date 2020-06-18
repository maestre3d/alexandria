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

func provideContext() context.Context {
	return Ctx
}

func InjectUserUseCase() (*interactor.UserUseCase, error) {
	wire.Build(
		provideContext,
		config.NewKernel,
		logger.NewZapLogger,
		wire.Bind(new(domain.UserRepository), new(*infrastructure.UserCognitoRepository)),
		infrastructure.NewUserCognitoRepository,
		interactor.NewUserUseCase,
	)

	return &interactor.UserUseCase{}, nil
}

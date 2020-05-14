// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

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

// Injectors from wire.go:

func InjectUserUseCase() (*interactor.UserUseCase, func(), error) {
	context := provideContext()
	logLogger := logger.NewZapLogger()
	kernelConfiguration, err := config.NewKernelConfiguration(context)
	if err != nil {
		return nil, nil, err
	}
	db, cleanup, err := persistence.NewPostgresPool(context, kernelConfiguration)
	if err != nil {
		return nil, nil, err
	}
	client, cleanup2, err := persistence.NewRedisPool(kernelConfiguration)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	userPostgresRepository := infrastructure.NewUserPostgresRepository(context, logLogger, db, client)
	interactorUserUseCase := interactor.NewUserUseCase(context, logLogger, userPostgresRepository)
	return interactorUserUseCase, func() {
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var persistenceSet = wire.NewSet(
	provideContext, config.NewKernelConfiguration, persistence.NewPostgresPool, persistence.NewRedisPool,
)

var userRepository = wire.NewSet(logger.NewZapLogger, persistenceSet, wire.Bind(new(domain.UserRepository), new(*infrastructure.UserPostgresRepository)), infrastructure.NewUserPostgresRepository)

var userUseCase = wire.NewSet(
	userRepository, interactor.NewUserUseCase,
)

func provideContext() context.Context {
	return context.Background()
}

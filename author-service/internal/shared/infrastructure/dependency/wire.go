// +build wireinject

package dependency

import (
	"context"
	"go.uber.org/zap/zapcore"

	"github.com/go-kit/kit/log"
	logZap "github.com/go-kit/kit/log/zap"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/author/infrastructure"
	"github.com/maestre3d/alexandria/author-service/internal/author/interactor"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/persistence"
	"go.uber.org/zap"
)

var configSet = wire.NewSet(
	ProvideContext,
	ProvideLogger,
	config.NewKernelConfig,
)

var DBMSPoolSet = wire.NewSet(
	configSet,
	persistence.NewPostgresPool,
)

var AuthorDBMSRepositorySet = wire.NewSet(
	DBMSPoolSet,
	persistence.NewRedisPool,
	wire.Bind(new(domain.IAuthorRepository), new(*infrastructure.AuthorDBMSRepository)),
	infrastructure.NewAuthorDBMSRepository,
)

var AuthorUseCaseSet = wire.NewSet(
	AuthorDBMSRepositorySet,
	wire.Bind(new(domain.IAuthorEventBus), new(*infrastructure.AuthorAWSEventBus)),
	infrastructure.NewAuthorAWSEventBus,
	interactor.NewAuthorUseCase,
)

func ProvideContext() context.Context {
	return context.Background()
}

func ProvideLogger() log.Logger {
	loggerZap, _ := zap.NewProduction()
	defer loggerZap.Sync()
	level := zapcore.Level(8)

	return logZap.NewZapSugarLogger(loggerZap, level)
}

func InjectAuthorUseCase() (*interactor.AuthorUseCase, func(), error) {
	wire.Build(AuthorUseCaseSet)

	return &interactor.AuthorUseCase{}, nil, nil
}

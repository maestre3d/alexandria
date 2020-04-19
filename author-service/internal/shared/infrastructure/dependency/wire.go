// +build wireinject

package dependency

import (
	"context"
	"go.uber.org/zap/zapcore"

	"github.com/go-kit/kit/log"
	logZap "github.com/go-kit/kit/log/zap"
	"go.uber.org/zap"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/author/infrastructure"
	"github.com/maestre3d/alexandria/author-service/internal/author/interactor"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/persistence"
)

/*
var LoggerSet = wire.NewSet(
	wire.Bind(new(util.ILogger) , new(*logging.ZapLogger)),
	logging.NewZapLogger,
)

 */

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

var AuthorServiceSet = wire.NewSet(
	AuthorDBMSRepositorySet,
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
	wire.Build(AuthorServiceSet)

	return &interactor.AuthorUseCase{}, nil, nil
}
// +build wireinject

package dependency

import (
	"context"

	"github.com/go-kit/kit/log"
	logZap "github.com/go-kit/kit/log/zap"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/domain"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/infrastructure"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/interactor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var configSet = wire.NewSet(
	provideContext,
	provideZapLogger,
	config.NewKernelConfig,
)

var watcherDynamoRepositorySet = wire.NewSet(
	configSet,
	wire.Bind(new(domain.WatcherRepository), new(*infrastructure.WatcherDynamoRepository)),
	infrastructure.NewWatcherDynamoRepository,
)

var watcherUseCaseSet = wire.NewSet(
	watcherDynamoRepositorySet,
	interactor.NewWatcherUseCase,
)

func provideContext() context.Context {
	return context.Background()
}

func provideZapLogger() log.Logger {
	loggerZap, _ := zap.NewProduction()
	defer loggerZap.Sync()
	level := zapcore.Level(8)

	return logZap.NewZapSugarLogger(loggerZap, level)
}

func InjectWatcherUseCase() (*interactor.WatcherUseCase, func(), error) {
	wire.Build(watcherUseCaseSet)

	return &interactor.WatcherUseCase{}, nil, nil
}

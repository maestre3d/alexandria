// +build wireinject

package dependency

import (
	"context"

	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/domain"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/infrastructure"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/interactor"
)

var configSet = wire.NewSet(
	provideContext,
	logger.NewZapLogger,
	config.NewKernelConfiguration,
)

var watcherDynamoRepositorySet = wire.NewSet(
	configSet,
	persistence.NewDynamoDBCollectionPool,
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

func InjectWatcherUseCase() (*interactor.WatcherUseCase, func(), error) {
	wire.Build(watcherUseCaseSet)

	return &interactor.WatcherUseCase{}, nil, nil
}

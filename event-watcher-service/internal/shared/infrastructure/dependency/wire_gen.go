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
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/domain"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/infrastructure"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/interactor"
)

// Injectors from wire.go:

func InjectEventUseCase() (*interactor.EventUseCase, func(), error) {
	context := provideContext()
	logLogger := logger.NewZapLogger()
	kernelConfiguration, err := config.NewKernelConfiguration(context)
	if err != nil {
		return nil, nil, err
	}
	collection, cleanup, err := persistence.NewDynamoDBCollectionPool(context, kernelConfiguration)
	if err != nil {
		return nil, nil, err
	}
	eventDynamoRepository := infrastructure.NewEventDynamoRepository(context, logLogger, collection)
	eventUseCase := interactor.NewEventUseCase(context, logLogger, eventDynamoRepository)
	return eventUseCase, func() {
		cleanup()
	}, nil
}

// wire.go:

var configSet = wire.NewSet(
	provideContext, logger.NewZapLogger, config.NewKernelConfiguration,
)

var eventDynamoRepositorySet = wire.NewSet(
	configSet, persistence.NewDynamoDBCollectionPool, wire.Bind(new(domain.EventRepository), new(*infrastructure.EventDynamoRepository)), infrastructure.NewEventDynamoRepository,
)

var eventUseCaseSet = wire.NewSet(
	eventDynamoRepositorySet, interactor.NewEventUseCase,
)

func provideContext() context.Context {
	return context.Background()
}
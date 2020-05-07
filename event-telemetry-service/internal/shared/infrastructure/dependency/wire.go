// +build wireinject

package dependency

import (
	"context"

	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/event-telemetry-service/internal/telemetry/domain"
	"github.com/maestre3d/alexandria/event-telemetry-service/internal/telemetry/infrastructure"
	"github.com/maestre3d/alexandria/event-telemetry-service/internal/telemetry/interactor"
)

var configSet = wire.NewSet(
	provideContext,
	logger.NewZapLogger,
	config.NewKernelConfiguration,
)

var eventDynamoRepositorySet = wire.NewSet(
	configSet,
	persistence.NewDynamoDBCollectionPool,
	wire.Bind(new(domain.EventRepository), new(*infrastructure.EventDynamoRepository)),
	infrastructure.NewEventDynamoRepository,
)

var eventUseCaseSet = wire.NewSet(
	eventDynamoRepositorySet,
	interactor.NewEventUseCase,
)

func provideContext() context.Context {
	return context.Background()
}

func InjectEventUseCase() (*interactor.EventUseCase, func(), error) {
	wire.Build(eventUseCaseSet)

	return &interactor.EventUseCase{}, nil, nil
}

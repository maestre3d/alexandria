// +build wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/maestre3d/alexandria/media-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/media-service/internal/interactor"
)

var Ctx = context.Background()

var dataSet = wire.NewSet(
	provideContext,
	config.NewKernel,
	persistence.NewPostgresPool,
	persistence.NewRedisPool,
	logger.NewZapLogger,
	wire.Bind(new(domain.MediaRepository), new(*infrastructure.MediaPQRepository)),
	infrastructure.NewMediaPQRepository,
)

var eventSet = wire.NewSet(
	wire.Bind(new(domain.MediaEvent), new(*infrastructure.MediaKafkaEvent)),
	infrastructure.NewMediaKafakaEvent,
)

func provideContext() context.Context {
	return Ctx
}

func InjectMediaUseCase() (*interactor.Media, func(), error) {
	wire.Build(dataSet, eventSet, interactor.NewMedia)
	return &interactor.Media{}, nil, nil
}

func InjectMediaSAGAUseCase() (*interactor.MediaSAGA, func(), error) {
	wire.Build(
		dataSet,
		eventSet,
		wire.Bind(new(domain.MediaEventSAGA), new(*infrastructure.MediaSAGAKafkaEvent)),
		infrastructure.NewMediaSAGAKafkaEvent,
		interactor.NewMediaSAGA,
	)
	return &interactor.MediaSAGA{}, nil, nil
}

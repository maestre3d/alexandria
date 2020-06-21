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
)

var mediaSet = wire.NewSet(
	dataSet,
	logger.NewZapLogger,
	wire.Bind(new(domain.MediaRepository), new(*infrastructure.MediaPQRepository)),
	infrastructure.NewMediaPQRepository,
	wire.Bind(new(domain.MediaEvent), new(*infrastructure.MediaKafkaEvent)),
	infrastructure.NewMediaKafakaEvent,
	interactor.NewMediaUseCase,
)

func provideContext() context.Context {
	return Ctx
}

func InjectMediaUseCase() (*interactor.MediaUseCase, func(), error) {
	wire.Build(mediaSet)
	return &interactor.MediaUseCase{}, nil, nil
}

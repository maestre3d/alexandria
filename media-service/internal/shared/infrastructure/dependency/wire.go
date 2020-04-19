// +build wireinject

package dependency

import (
	"context"

	"github.com/go-kit/kit/log"
	logZap "github.com/go-kit/kit/log/zap"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/media/infrastructure"
	"github.com/maestre3d/alexandria/media-service/internal/media/interactor"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/persistence"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var configSet = wire.NewSet(
	ProvideContext,
	ProvideLogger,
	config.NewKernelConfig,
)
var postgresPoolSet = wire.NewSet(
	configSet,
	persistence.NewPostgresPool,
)
var mediaRepositorySet = wire.NewSet(
	postgresPoolSet,
	persistence.NewRedisPool,
	infrastructure.NewMediaRDBMSRepository,
	wire.Bind(new(domain.IMediaRepository), new(*infrastructure.MediaDBMSRepository)),
)
var mediaUseCaseSet = wire.NewSet(
	mediaRepositorySet,
	interactor.NewMediaUseCase,
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

func InjectMediaUseCase() (*interactor.MediaUseCase, func(), error) {
	wire.Build(mediaUseCaseSet)

	return &interactor.MediaUseCase{}, nil, nil
}

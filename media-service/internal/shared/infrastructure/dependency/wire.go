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
	provideContext,
	provideLogger,
	config.NewKernelConfig,
)
var dBMSPoolSet = wire.NewSet(
	configSet,
	persistence.NewPostgresPool,
)
var mediaDBMSRepositorySet = wire.NewSet(
	dBMSPoolSet,
	persistence.NewRedisPool,
	wire.Bind(new(domain.IMediaRepository), new(*infrastructure.MediaDBMSRepository)),
	infrastructure.NewMediaDBMSRepository,
)

// Inject media's interactor EventBus
var mediaUseCaseSet = wire.NewSet(
	mediaDBMSRepositorySet,
	interactor.NewMediaUseCase,
)

func provideContext() context.Context {
	return context.Background()
}

func provideLogger() log.Logger {
	loggerZap, _ := zap.NewProduction()
	defer loggerZap.Sync()
	level := zapcore.Level(8)

	return logZap.NewZapSugarLogger(loggerZap, level)
}

func InjectMediaUseCase() (*interactor.MediaUseCase, func(), error) {
	wire.Build(mediaUseCaseSet)

	return &interactor.MediaUseCase{}, nil, nil
}

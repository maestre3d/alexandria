// +build wireinject

package dependency

import (
	"context"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/application"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/infrastructure"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/infrastructure/logging"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/infrastructure/persistence"
	"github.com/maestre3d/alexandria/src/media-service/pkg/service/delivery"
	"github.com/maestre3d/alexandria/src/media-service/pkg/service/delivery/handler"
)

var loggerSet = wire.NewSet(
	logging.NewLogger,
	wire.Bind(new(util.ILogger), new(*logging.Logger)),
)
var postgresPoolSet = wire.NewSet(
	ProvideContext,
	loggerSet,
	persistence.NewPostgresPool,
)
var mediaRepository = wire.NewSet(
	postgresPoolSet,
	infrastructure.NewMediaRDBMSRepository,
	wire.Bind(new(domain.IMediaRepository), new(*infrastructure.MediaRDBMSRepository)),
)
var mediaUseCaseSet = wire.NewSet(
	mediaRepository,
	application.NewMediaUseCase,
)
var mediaHandlerSet = wire.NewSet(
	mediaUseCaseSet,
	handler.NewMediaHandler,
)
var proxyHandlersSet = wire.NewSet(
	mediaHandlerSet,
	ProvideProxyHandlers,
)

func ProvideContext() context.Context {
	return context.Background()
}

func ProvideMediaLocalRepository(logger util.ILogger) *infrastructure.MediaLocalRepository {
	return infrastructure.NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), logger)
}

func ProvideProxyHandlers(media *handler.MediaHandler) *delivery.ProxyHandlers {
	// Map handlers to proxy
	return &delivery.ProxyHandlers{
		media,
	}
}

func InitHTTPServiceProxy() (*delivery.HTTPServiceProxy, func(), error) {
	wire.Build(wire.NewSet(
		proxyHandlersSet,
		delivery.NewHTTPServer,
		delivery.NewHTTPServiceProxy,
	))

	return &delivery.HTTPServiceProxy{}, nil, nil
}

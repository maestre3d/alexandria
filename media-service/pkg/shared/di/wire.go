// +build wireinject

package di

import (
	"context"
	"github.com/go-kit/kit/log"
	logzap "github.com/go-kit/kit/log/zap"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/media-service/internal/shared/infrastructure/dependency"
	"github.com/maestre3d/alexandria/media-service/pkg/media"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
	"github.com/maestre3d/alexandria/media-service/pkg/shared"
	"github.com/maestre3d/alexandria/media-service/pkg/transport"
	"github.com/maestre3d/alexandria/media-service/pkg/transport/handler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var mediaServiceSet = wire.NewSet(
	ProvideLogger,
	ProvideAuthorService,
)

var proxyHandlersSet = wire.NewSet(
	mediaServiceSet,
	handler.NewMediaHandler,
	ProvideProxyHandlers,
)

var httpProxySet = wire.NewSet(
	proxyHandlersSet,
	ProvideContext,
	config.NewKernelConfig,
	shared.NewHTTPServer,
	transport.NewHTTPTransportProxy,
)

func ProvideContext() context.Context {
	return context.Background()
}

func ProvideLogger() log.Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	level := zapcore.Level(8)

	return logzap.NewZapSugarLogger(zapLogger, level)
}

func ProvideAuthorService(logger log.Logger) (service.IMediaService, func(), error) {
	authorUseCase, cleanup, err := dependency.InjectMediaUseCase()

	authorService := media.NewMediaService(authorUseCase, logger)

	return authorService, cleanup, err
}

func ProvideProxyHandlers(mediaHandler *handler.MediaHandler) *transport.ProxyHandlers {
	return &transport.ProxyHandlers{mediaHandler}
}

func InjectHTTPProxy() (*transport.HTTPTransportProxy, func(), error) {
	wire.Build(
		httpProxySet,
	)

	return &transport.HTTPTransportProxy{}, nil, nil
}

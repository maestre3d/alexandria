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
	"github.com/maestre3d/alexandria/media-service/pkg/transport/pb"
	"github.com/maestre3d/alexandria/media-service/pkg/transport/proxy"
	"github.com/maestre3d/alexandria/media-service/pkg/transport/tracer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var mediaServiceSet = wire.NewSet(
	provideLogger,
	provideMediaService,
)

var proxyHandlersSet = wire.NewSet(
	mediaServiceSet,
	config.NewKernelConfig,
	tracer.NewOpenTracer,
	tracer.NewZipkinTracer,
	handler.NewMediaHandler,
	provideProxyHandlers,
)

var httpProxySet = wire.NewSet(
	proxyHandlersSet,
	provideContext,
	shared.NewHTTPServer,
	proxy.NewHTTPTransportProxy,
)

var rpcProxyHandlersSet = wire.NewSet(
	handler.NewMediaRPCServer,
	provideRPCProxyHandlers,
)

var rpcProxySet = wire.NewSet(
	rpcProxyHandlersSet,
	proxy.NewRPCTransportProxy,
)

func provideContext() context.Context {
	return context.Background()
}

func provideLogger() log.Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	level := zapcore.Level(8)

	return logzap.NewZapSugarLogger(zapLogger, level)
}

func provideMediaService(logger log.Logger) (service.IMediaService, func(), error) {
	mediaUseCase, cleanup, err := dependency.InjectMediaUseCase()

	mediaService := media.NewMediaService(mediaUseCase, logger)

	return mediaService, cleanup, err
}

func provideProxyHandlers(mediaHandler *handler.MediaHandler) *proxy.ProxyHandlers {
	return &proxy.ProxyHandlers{mediaHandler}
}

func provideRPCProxyHandlers(mediaHandler pb.MediaServer) *proxy.RPCProxyHandlers {
	return &proxy.RPCProxyHandlers{mediaHandler}
}

func InjectTransportService() (*transport.TransportService, func(), error) {
	wire.Build(httpProxySet, rpcProxySet, transport.NewTransportService)

	return &transport.TransportService{}, nil, nil
}


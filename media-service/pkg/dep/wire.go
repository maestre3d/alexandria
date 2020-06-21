// +build wireinject

package dep

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/tracer"
	"github.com/alexandria-oss/core/transport"
	"github.com/alexandria-oss/core/transport/proxy"
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/media-service/internal/dependency"
	"github.com/maestre3d/alexandria/media-service/pkg/media"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	"github.com/maestre3d/alexandria/media-service/pkg/transport/bind"
)

var Ctx = context.Background()

var interactorSet = wire.NewSet(
	provideContext,
	logger.NewZapLogger,
	provideMediaInteractor,
)

var httpProxySet = wire.NewSet(
	interactorSet,
	config.NewKernel,
	tracer.NewZipkin,
	tracer.WrapZipkinOpenTracing,
	bind.NewMediaHTTP,
	provideHTTPHandlers,
	proxy.NewHTTP,
)

var rpcProxySet = wire.NewSet(
	bind.NewMediaRPC,
	bind.NewHealthRPC,
	provideRPCServers,
	proxy.NewRPC,
)

var eventProxySet = wire.NewSet(
	provideMediaSAGAInteractor,
	bind.NewMediaEventConsumer,
	provideEventConsumers,
	proxy.NewEvent,
)

func provideContext() context.Context {
	return Ctx
}

func provideMediaInteractor(ctx context.Context, logger log.Logger) (usecase.MediaInteractor, func(), error) {
	dependency.Ctx = ctx

	mediaInteractor, cleanup, err := dependency.InjectMediaUseCase()
	mediaService := media.WrapMediaInstrumentation(mediaInteractor, logger)

	return mediaService, cleanup, err
}

func provideMediaSAGAInteractor(ctx context.Context, logger log.Logger) (usecase.MediaSAGAInteractor, func(), error) {
	dependency.Ctx = ctx

	mediaInteractor, cleanup, err := dependency.InjectMediaSAGAUseCase()
	mediaService := media.WrapMediaSAGAInstrumentation(mediaInteractor, logger)

	return mediaService, cleanup, err
}

// Bind/Map used http handlers
func provideHTTPHandlers(mediaHandler *bind.MediaHandler) []proxy.Handler {
	handlers := make([]proxy.Handler, 0)
	handlers = append(handlers, mediaHandler)
	return handlers
}

// Bind/Map used rpc servers
func provideRPCServers(mediaServer *bind.MediaRPCServer, healthServer *bind.HealthRPCServer) []proxy.RPCServer {
	servers := make([]proxy.RPCServer, 0)
	servers = append(servers, mediaServer, healthServer)
	return servers
}

// Bind/Map used event consumers
func provideEventConsumers(mediaConsumer *bind.MediaEventConsumer) []proxy.Consumer {
	consumers := make([]proxy.Consumer, 0)
	consumers = append(consumers, mediaConsumer)
	return consumers
}

func InjectTransportService() (*transport.Transport, func(), error) {
	wire.Build(httpProxySet, rpcProxySet, eventProxySet, transport.NewTransport)

	return &transport.Transport{}, nil, nil
}

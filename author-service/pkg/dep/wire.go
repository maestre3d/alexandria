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
	"github.com/maestre3d/alexandria/author-service/internal/dependency"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/bind"
)

var Ctx context.Context = context.Background()

var authorInteractorSet = wire.NewSet(
	logger.NewZapLogger,
	provideAuthorInteractor,
)

var httpProxySet = wire.NewSet(
	authorInteractorSet,
	provideContext,
	config.NewKernel,
	tracer.NewZipkin,
	tracer.WrapZipkinOpenTracing,
	bind.NewAuthorHTTP,
	provideHTTPHandlers,
	proxy.NewHTTP,
)

var rpcProxySet = wire.NewSet(
	bind.NewAuthorRPC,
	bind.NewHealthRPC,
	provideRPCServers,
	proxy.NewRPC,
)

var eventProxySet = wire.NewSet(
	provideAuthorSAGAInteractor,
	bind.NewAuthorEventConsumer,
	provideEventConsumers,
	proxy.NewEvent,
)

func provideContext() context.Context {
	return Ctx
}

func provideAuthorInteractor(logger log.Logger) (usecase.AuthorInteractor, func(), error) {
	dependency.Ctx = Ctx
	authorUseCase, cleanup, err := dependency.InjectAuthorUseCase()

	authorService := author.WrapAuthorInstrumentation(authorUseCase, logger)

	return authorService, cleanup, err
}

func provideAuthorSAGAInteractor(logger log.Logger) (usecase.AuthorSAGAInteractor, func(), error) {
	dependency.Ctx = Ctx
	authorUseCase, cleanup, err := dependency.InjectAuthorSAGAUseCase()

	authorService := author.WrapAuthorSAGAInstrumentation(authorUseCase, logger)

	return authorService, cleanup, err
}

// Bind/Map used http handlers
func provideHTTPHandlers(authorHandler *bind.AuthorHandler) []proxy.Handler {
	handlers := make([]proxy.Handler, 0)
	handlers = append(handlers, authorHandler)
	return handlers
}

// Bind/Map used rpc servers
func provideRPCServers(authorServer *bind.AuthorRPCServer, healthServer *bind.HealthRPCServer) []proxy.RPCServer {
	servers := make([]proxy.RPCServer, 0)
	servers = append(servers, authorServer)
	servers = append(servers, healthServer)
	return servers
}

func provideEventConsumers(authorHandler *bind.AuthorEventConsumer) []proxy.Consumer {
	consumers := make([]proxy.Consumer, 0)
	consumers = append(consumers, authorHandler)
	return consumers
}

func InjectTransportService() (*transport.Transport, func(), error) {
	wire.Build(httpProxySet, rpcProxySet, eventProxySet, transport.NewTransport)

	return &transport.Transport{}, nil, nil
}

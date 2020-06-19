// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

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

// Injectors from wire.go:

func InjectTransportService() (*transport.Transport, func(), error) {
	logLogger := logger.NewZapLogger()
	authorInteractor, cleanup, err := provideAuthorInteractor(logLogger)
	if err != nil {
		return nil, nil, err
	}
	context := provideContext()
	kernel, err := config.NewKernel(context)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	zipkinTracer, cleanup2 := tracer.NewZipkin(kernel)
	opentracingTracer := tracer.WrapZipkinOpenTracing(kernel, zipkinTracer)
	authorRPCServer := bind.NewAuthorRPC(authorInteractor, logLogger, opentracingTracer, zipkinTracer)
	healthRPCServer := bind.NewHealthRPC()
	v := provideRPCServers(authorRPCServer, healthRPCServer)
	server, cleanup3 := proxy.NewRPC(v)
	authorHandler := bind.NewAuthorHTTP(authorInteractor, logLogger, opentracingTracer, zipkinTracer)
	v2 := provideHTTPHandlers(authorHandler)
	http, cleanup4 := proxy.NewHTTP(kernel, v2...)
	authorEventConsumer := bind.NewAuthorEventConsumer(authorInteractor, logLogger)
	v3 := provideEventConsumers(authorEventConsumer)
	event, cleanup5, err := proxy.NewEvent(context, kernel, v3...)
	if err != nil {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	transportTransport := transport.NewTransport(server, http, event, kernel)
	return transportTransport, func() {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var Ctx context.Context = context.Background()

var authorInteractorSet = wire.NewSet(logger.NewZapLogger, provideAuthorInteractor)

var httpProxySet = wire.NewSet(
	authorInteractorSet,
	provideContext, config.NewKernel, tracer.NewZipkin, tracer.WrapZipkinOpenTracing, bind.NewAuthorHTTP, provideHTTPHandlers, proxy.NewHTTP,
)

var rpcProxySet = wire.NewSet(bind.NewAuthorRPC, bind.NewHealthRPC, provideRPCServers, proxy.NewRPC)

var eventProxySet = wire.NewSet(bind.NewAuthorEventConsumer, provideEventConsumers, proxy.NewEvent)

func provideContext() context.Context {
	return Ctx
}

func provideAuthorInteractor(logger2 log.Logger) (usecase.AuthorInteractor, func(), error) {
	dependency.Ctx = Ctx
	authorUseCase, cleanup, err := dependency.InjectAuthorUseCase()

	authorService := author.WrapAuthorInstrumentation(authorUseCase, logger2)

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

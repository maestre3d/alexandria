// +build wireinject

package dep

import (
	"context"
	oczipkin "contrib.go.opencensus.io/exporter/zipkin"
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
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

var Ctx context.Context = context.Background()

var authorInteractorSet = wire.NewSet(
	logger.NewZapLogger,
	provideAuthorInteractor,
)

var zipkinSet = wire.NewSet(
	provideZipkinReporter,
	provideZipkinEndpoint,
	provideZipkinTracer,
)

var httpProxySet = wire.NewSet(
	authorInteractorSet,
	provideContext,
	config.NewKernel,
	zipkinSet,
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

/* ZIPKIN PROVIDERS */

// NewZipkin returns a zipkin tracing consumer
func provideZipkinReporter(cfg *config.Kernel) (reporter.Reporter, func()) {
	if cfg.Tracing.ZipkinHost != "" && cfg.Tracing.ZipkinEndpoint != "" {
		zipkinReporter := zipkinhttp.NewReporter(cfg.Tracing.ZipkinHost)
		cleanup := func() {
			_ = zipkinReporter.Close()
		}

		return zipkinReporter, cleanup
	}

	return nil, nil
}

// NewZipkin returns a zipkin tracing consumer
func provideZipkinEndpoint(cfg *config.Kernel) *model.Endpoint {
	if cfg.Tracing.ZipkinHost != "" && cfg.Tracing.ZipkinEndpoint != "" {
		zipkinEndpoint, err := zipkin.NewEndpoint(cfg.Service, cfg.Tracing.ZipkinEndpoint)
		if err != nil {
			return nil
		}

		return zipkinEndpoint
	}

	return nil
}

// NewZipkin returns a zipkin tracing consumer
func provideZipkinTracer(cfg *config.Kernel, r reporter.Reporter, ep *model.Endpoint) *zipkin.Tracer {
	if cfg.Tracing.ZipkinHost != "" && cfg.Tracing.ZipkinEndpoint != "" {
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		trace.RegisterExporter(oczipkin.NewExporter(r, ep))

		zipkinTrace, err := zipkin.NewTracer(r, zipkin.WithLocalEndpoint(ep))
		if err != nil {
			return nil
		}

		return zipkinTrace
	}

	return nil
}

func InjectTransportService() (*transport.Transport, func(), error) {
	wire.Build(httpProxySet, rpcProxySet, eventProxySet, transport.NewTransport)

	return &transport.Transport{}, nil, nil
}

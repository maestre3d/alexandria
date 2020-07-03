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
	"github.com/maestre3d/alexandria/blob-service/internal/dependency"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	"github.com/maestre3d/alexandria/blob-service/pkg/transport/bind"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

var Ctx = context.Background()

var interactorSet = wire.NewSet(
	provideContext,
	logger.NewZapLogger,
	provideBlobInteractor,
)

var zipkinSet = wire.NewSet(
	provideZipkinReporter,
	provideZipkinEndpoint,
	provideZipkinTracer,
)

var httpProxySet = wire.NewSet(
	interactorSet,
	config.NewKernel,
	zipkinSet,
	tracer.WrapZipkinOpenTracing,
	bind.NewBlobHandler,
	provideHTTPHandlers,
	proxy.NewHTTP,
)

var eventProxySet = wire.NewSet(
	bind.NewBlobEventConsumer,
	provideEventConsumers,
	proxy.NewEvent,
)

func provideContext() context.Context {
	return Ctx
}

func provideBlobInteractor(log log.Logger) (usecase.BlobInteractor, func(), error) {
	dependency.Ctx = Ctx

	interactor, cleanup, err := dependency.InjectBlobUseCase()
	svc := blob.WrapBlobInstrumentation(interactor, log)

	return svc, cleanup, err
}

// Bind/Map used http handlers
func provideHTTPHandlers(blobHandler *bind.BlobHandler) []proxy.Handler {
	handlers := make([]proxy.Handler, 0)
	handlers = append(handlers, blobHandler)
	return handlers
}

// Bind/Map used rpc servers
func provideRPCServers() []proxy.RPCServer {
	servers := make([]proxy.RPCServer, 0)
	// servers = append(servers, fooServer, healthServer)
	return servers
}

// Bind/Map used event consumers
func provideEventConsumers(blobConsumer *bind.BlobEventConsumer) []proxy.Consumer {
	consumers := make([]proxy.Consumer, 0)
	consumers = append(consumers, blobConsumer)
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
	wire.Build(httpProxySet, provideRPCServers, proxy.NewRPC, eventProxySet, transport.NewTransport)

	return &transport.Transport{}, nil, nil
}

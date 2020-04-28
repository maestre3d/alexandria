package tracer

import (
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/config"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func NewZipkinTracer(logger log.Logger, cfg *config.KernelConfig) (*zipkin.Tracer, func()) {
	if cfg.MetricConfig.ZipkinHost != "" && cfg.MetricConfig.ZipkinEndpoint != "" {
		zipkinReporter := zipkinhttp.NewReporter(cfg.MetricConfig.ZipkinHost)
		cleanup := func() {
			zipkinReporter.Close()
		}

		zipkinEndpoint, err := zipkin.NewEndpoint(cfg.Service, cfg.MetricConfig.ZipkinEndpoint)
		if err != nil {
			logger.Log("method", "public.infrastructure.transport.tracing", "err", err.Error())
			return nil, cleanup
		}

		zipkinTrace, err := zipkin.NewTracer(zipkinReporter, zipkin.WithLocalEndpoint(zipkinEndpoint))
		if err != nil {
			logger.Log("method", "public.infrastructure.transport.tracing", "err", err.Error())
			return nil, cleanup
		}

		return zipkinTrace, cleanup
	}

	return nil, nil
}

func NewOpenTracer(logger log.Logger, cfg *config.KernelConfig, zipTracer *zipkin.Tracer) stdopentracing.Tracer {
	// Using zipkin with OpenCensus instrumentation
	if cfg.MetricConfig.ZipkinBridge && zipTracer != nil {
		logger.Log("tracer", "zipkin", "type", "opentracing", "url", cfg.MetricConfig.ZipkinHost)
		return zipkinot.Wrap(zipTracer)
	}

	// Return Non-zipkin tracing
	return stdopentracing.GlobalTracer()
}

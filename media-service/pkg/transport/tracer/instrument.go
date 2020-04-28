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
	if cfg.MetricConfig.ZipkinHost != "" {
		reporter := zipkinhttp.NewReporter(cfg.MetricConfig.ZipkinHost)
		zEP, _ := zipkin.NewEndpoint(cfg.Service, cfg.MetricConfig.ZipkinEndpoint)
		zipkinTrace, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
		if err != nil {
			logger.Log("err", "failed to start zipkin tracer", "resource", "public.infrastructure.transport.metrics")
		}

		cleanup := func() {
			err = reporter.Close()
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

	// No operation
	return stdopentracing.GlobalTracer()
}

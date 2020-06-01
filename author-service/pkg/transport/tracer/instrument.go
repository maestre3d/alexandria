package tracer

import (
	"github.com/alexandria-oss/core/config"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func NewZipkinTracer(cfg *config.Kernel) (*zipkin.Tracer, func()) {
	if cfg.Tracing.ZipkinHost != "" && cfg.Tracing.ZipkinEndpoint != "" {
		zipkinReporter := zipkinhttp.NewReporter(cfg.Tracing.ZipkinHost)
		cleanup := func() {
			zipkinReporter.Close()
		}

		zipkinEndpoint, err := zipkin.NewEndpoint(cfg.Service, cfg.Tracing.ZipkinEndpoint)
		if err != nil {
			return nil, cleanup
		}

		zipkinTrace, err := zipkin.NewTracer(zipkinReporter, zipkin.WithLocalEndpoint(zipkinEndpoint))
		if err != nil {
			return nil, cleanup
		}

		return zipkinTrace, cleanup
	}

	return nil, nil
}

func WrapOpenTracer(cfg *config.Kernel, zipTracer *zipkin.Tracer) stdopentracing.Tracer {
	// Using zipkin with OpenTracing middleware
	if cfg.Tracing.ZipkinBridge && zipTracer != nil {
		return zipkinot.Wrap(zipTracer)
	}

	// Return Non-zipkin tracing
	return stdopentracing.GlobalTracer()
}

package shared

import (
	"fmt"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

// WrapResiliency inject fault-tolerant/resiliency patterns into the given endpoint
func WrapResiliency(e endpoint.Endpoint, service, action string) endpoint.Endpoint {
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          fmt.Sprintf("%s.%s", service, action),
		MaxRequests:   100,
		Interval:      0,
		Timeout:       15 * time.Second,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})
	e = ratelimit.NewErroringLimiter(limiter)(e)
	return circuitbreaker.Gobreaker(cb)(e)
}

type WrapInstrumentParams struct {
	Logger       log.Logger
	Duration     metrics.Histogram
	Tracer       stdopentracing.Tracer
	ZipkinTracer *stdzipkin.Tracer
}

// WrapInstrumentation inject basic instrumentation into the given endpoint
func WrapInstrumentation(e endpoint.Endpoint, service, action string, params *WrapInstrumentParams) endpoint.Endpoint {
	// Transport instrumentation
	// Distributed Tracing
	// OpenTracing - Register the service endpoint to get OpenTracing's standards (like request_id)
	e = kitoc.TraceEndpoint("gokit:endpoint " + action)(e)
	// OpenTracing server
	e = opentracing.TraceServer(params.Tracer, action)(e)
	// TODO: Optional - Add Jaeger instead Zipkin tracer consumer
	if params.ZipkinTracer != nil {
		e = zipkin.TraceEndpoint(params.ZipkinTracer, action)(e)
	}

	// Wrap this endpoint with the required instrumentation left
	// using a chain of responsibility pattern
	// Using zap logger
	location := fmt.Sprintf("%s.%s", service, action)
	e = LoggingMiddleware(log.With(params.Logger, "method", location))(e)
	// Using prometheus + OpenCensus
	return MetricMiddleware(params.Duration.With("method", location))(e)
}

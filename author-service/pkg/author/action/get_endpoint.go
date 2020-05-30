package action

import (
	"context"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

type GetRequest struct {
	ID string `json:"id"`
}

type GetResponse struct {
	Author *domain.AuthorEntity `json:"author"`
	Err    error                `json:"-"`
}

func MakeGetAuthorEndpoint(svc usecase.IAuthorService, logger log.Logger, duration metrics.Histogram, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		author, err := svc.Get(req.ID)
		if err != nil {
			return GetResponse{
				Author: nil,
				Err:    err,
			}, nil
		}

		return GetResponse{
			Author: author,
			Err:    nil,
		}, nil
	}

	// Transport's fault-tolerant patterns
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "author.get",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       15 * time.Second,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})
	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)

	// Distributed Tracing
	// OpenCensus tracer
	ep = kitoc.TraceEndpoint("gokit:endpoint get")(ep)
	// OpenTracing server
	ep = opentracing.TraceServer(tracer, "Get")(ep)
	if zipkinTracer != nil {
		ep = zipkin.TraceEndpoint(zipkinTracer, "Get")(ep)
	}

	// Transport metrics
	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.get"))(ep)
	ep = shared.InstrumentingMiddleware(duration.With("method", "author.get"))(ep)

	return ep
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = GetResponse{}
)

func (r GetResponse) Failed() error { return r.Err }

package action

import (
	"context"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type CreateRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	BirthDate   string `json:"birth_date"`
	OwnerID     string `json:"owner_id"`
}

type CreateResponse struct {
	Author *domain.Author `json:"author"`
	Err    error          `json:"-"`
}

func MakeCreateAuthorEndpoint(svc usecase.AuthorInteractor, logger log.Logger, duration metrics.Histogram, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateRequest)

		aggr := &domain.AuthorAggregate{
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			DisplayName: req.DisplayName,
			OwnerID:     req.OwnerID,
			BirthDate:   req.BirthDate,
		}

		createdAuthor, err := svc.Create(ctx, aggr)
		if err != nil {
			return CreateResponse{
				Author: nil,
				Err:    err,
			}, err
		}

		return CreateResponse{
			Author: createdAuthor,
			Err:    nil,
		}, nil
	}

	// Transport's fault-tolerant patterns
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "author.create",
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
	ep = kitoc.TraceEndpoint("gokit:endpoint create")(ep)
	// OpenTracing server
	ep = opentracing.TraceServer(tracer, "Create")(ep)
	if zipkinTracer != nil {
		ep = zipkin.TraceEndpoint(zipkinTracer, "Create")(ep)
	}

	// Transport metrics
	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.create"))(ep)
	ep = shared.InstrumentingMiddleware(duration.With("method", "author.create"))(ep)

	return ep
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = CreateResponse{}
)

func (r CreateResponse) Failed() error { return r.Err }

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
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

type UpdateRequest struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	BirthDate   string `json:"birth_date"`
}

type UpdateResponse struct {
	Author *domain.AuthorEntity `json:"author"`
	Err    error                `json:"-"`
}

func MakeUpdateAuthorEndpoint(svc service.IAuthorService, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UpdateRequest)
		author, err := svc.Update(req.ID, req.FirstName, req.LastName, req.DisplayName, req.BirthDate)
		if err != nil {
			return UpdateResponse{
				Author: nil,
				Err:    err,
			}, nil
		}

		return UpdateResponse{
			Author: author,
			Err:    nil,
		}, nil
	}

	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "author.update",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       15 * time.Second,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})

	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)
	ep = kitoc.TraceEndpoint("gokit:endpoint update")(ep)
	ep = opentracing.TraceServer(otTracer, "Update")(ep)
	if zipkinTracer != nil {
		ep = zipkin.TraceEndpoint(zipkinTracer, "Update")(ep)
	}
	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.update"))(ep)
	ep = shared.InstrumentingMiddleware(duration.With("method", "author.update"))(ep)

	return ep
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = UpdateResponse{}
)

func (r UpdateResponse) Failed() error { return r.Err }

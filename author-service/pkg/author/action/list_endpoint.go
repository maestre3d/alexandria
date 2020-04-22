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
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

type ListRequest struct {
	PageToken    string            `json:"page_token"`
	PageSize     string            `json:"page_size"`
	FilterParams util.FilterParams `json:"filter_params"`
}

type ListResponse struct {
	Authors       []*domain.AuthorEntity `json:"authors"`
	NextPageToken string                 `json:"next_page_token"`
	Err           error                  `json:"-"`
}

func MakeListAuthorEndpoint(svc service.IAuthorService, logger log.Logger, duration metrics.Histogram, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ListRequest)
		authors, nextToken, err := svc.List(req.PageToken, req.PageSize, req.FilterParams)
		if err != nil {
			return ListResponse{
				Authors:       nil,
				NextPageToken: "",
				Err:           err,
			}, err
		}

		return ListResponse{
			Authors:       authors,
			NextPageToken: nextToken,
			Err:           nil,
		}, nil
	}

	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "author.list",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       15 * time.Second,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})

	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)
	ep = kitoc.TraceEndpoint("gokit:endpoint list")(ep)
	ep = opentracing.TraceServer(otTracer, "List")(ep)
	if zipkinTracer != nil {
		ep = zipkin.TraceEndpoint(zipkinTracer, "List")(ep)
	}
	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.list"))(ep)
	ep = shared.InstrumentingMiddleware(duration.With("method", "author.list"))(ep)

	return ep
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = ListResponse{}
)

func (r ListResponse) Failed() error { return r.Err }

package action

import (
	"context"
	"github.com/go-kit/kit/metrics"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
	"github.com/maestre3d/alexandria/media-service/pkg/shared"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type UpdateRequest struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	UserID      string `json:"user_id"`
	AuthorID    string `json:"author_id"`
	PublishDate string `json:"publish_date"`
	MediaType   string `json:"media_type"`
}

type UpdateResponse struct {
	Media *domain.MediaEntity `json:"media"`
	Err   error               `json:"-"`
}

func MakeUpdateMediaEndpoint(svc service.IMediaService, logger log.Logger, duration metrics.Histogram, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UpdateRequest)
		updatedMedia, err := svc.Update(req.ID, req.Title, req.DisplayName, req.Description, req.UserID, req.AuthorID, req.PublishDate, req.MediaType)
		if err != nil {
			return UpdateResponse{
				Media: nil,
				Err:   err,
			}, nil
		}

		return UpdateResponse{
			Media: updatedMedia,
			Err:   nil,
		}, nil
	}

	// Transport's fault-tolerant patterns
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "media.update",
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
	ep = kitoc.TraceEndpoint("gokit:endpoint update")(ep)
	// OpenTracing server
	ep = opentracing.TraceServer(tracer, "Update")(ep)
	if zipkinTracer != nil {
		ep = zipkin.TraceEndpoint(zipkinTracer, "Update")(ep)
	}

	// Transport metrics
	ep = shared.LoggingMiddleware(log.With(logger, "method", "media.update"))(ep)
	ep = shared.InstrumentingMiddleware(duration.With("method", "media.update"))(ep)

	return ep
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = UpdateResponse{}
)

func (r UpdateResponse) Failed() error { return r.Err }
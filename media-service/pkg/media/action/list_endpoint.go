package action

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/middleware"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type ListRequest struct {
	PageToken    string            `json:"page_token"`
	PageSize     string            `json:"page_size"`
	FilterParams core.FilterParams `json:"filter_params"`
}

type ListResponse struct {
	Medias        []*domain.Media `json:"media"`
	NextPageToken string          `json:"next_page_token"`
	Err           error           `json:"-"`
}

func MakeListMediaEndpoint(svc usecase.MediaInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ListRequest)
		Medias, nextToken, err := svc.List(ctx, req.PageToken, req.PageSize, req.FilterParams)
		if err != nil {
			return ListResponse{
				Medias:        nil,
				NextPageToken: "",
				Err:           err,
			}, err
		}

		return ListResponse{
			Medias:        Medias,
			NextPageToken: nextToken,
			Err:           nil,
		}, nil
	}

	// Required resiliency and instrumentation
	action := "list"
	ep = middleware.WrapResiliency(ep, "media", action)
	return middleware.WrapInstrumentation(ep, "media", action, &middleware.WrapInstrumentParams{
		Logger:       logger,
		Duration:     duration,
		Tracer:       tracer,
		ZipkinTracer: zipkinTracer,
	})
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = ListResponse{}
)

func (r ListResponse) Failed() error { return r.Err }

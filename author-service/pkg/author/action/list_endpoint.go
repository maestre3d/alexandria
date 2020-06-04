package action

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type ListRequest struct {
	PageToken    string            `json:"page_token"`
	PageSize     string            `json:"page_size"`
	FilterParams core.FilterParams `json:"filter_params"`
}

type ListResponse struct {
	Authors       []*domain.Author `json:"authors"`
	NextPageToken string           `json:"next_page_token"`
	Err           error            `json:"-"`
}

func MakeListAuthorEndpoint(svc usecase.AuthorInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ListRequest)
		authors, nextToken, err := svc.List(ctx, req.PageToken, req.PageSize, req.FilterParams)
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

	// Required resiliency and instrumentation
	action := "list"
	ep = shared.WrapResiliency(ep, "author", action)
	return shared.WrapInstrumentation(ep, "author", action, &shared.WrapInstrumentParams{
		logger,
		duration,
		tracer,
		zipkinTracer,
	})
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = ListResponse{}
)

func (r ListResponse) Failed() error { return r.Err }

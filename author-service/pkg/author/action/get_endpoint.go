package action

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type GetRequest struct {
	ID string `json:"id"`
}

type GetResponse struct {
	Author *domain.Author `json:"author"`
	Err    error          `json:"-"`
}

func MakeGetAuthorEndpoint(svc usecase.AuthorInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		author, err := svc.Get(ctx, req.ID)
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

	// Required resiliency and instrumentation
	action := "get"
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
	_ endpoint.Failer = GetResponse{}
)

func (r GetResponse) Failed() error { return r.Err }

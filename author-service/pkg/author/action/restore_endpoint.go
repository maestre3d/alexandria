package action

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type RestoreRequest struct {
	ID string `json:"id"`
}

type RestoreResponse struct {
	Err error `json:"-"`
}

func MakeRestoreAuthorEndpoint(svc usecase.AuthorInteractor, logger log.Logger, duration metrics.Histogram, tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(RestoreRequest)
		err = svc.Restore(ctx, req.ID)
		if err != nil {
			return RestoreResponse{err}, nil
		}

		return RestoreResponse{nil}, nil
	}

	// Required resiliency and instrumentation
	action := "restore"
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
	_ endpoint.Failer = RestoreResponse{}
)

func (r RestoreResponse) Failed() error { return r.Err }

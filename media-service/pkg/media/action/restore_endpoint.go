package action

import (
	"context"
	"github.com/alexandria-oss/core/middleware"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type RestoreRequest struct {
	ID string `json:"id"`
}

type RestoreResponse struct {
	Err error `json:"-"`
}

func MakeRestoreMediaEndpoint(svc usecase.MediaInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
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
	_ endpoint.Failer = RestoreResponse{}
)

func (r RestoreResponse) Failed() error { return r.Err }

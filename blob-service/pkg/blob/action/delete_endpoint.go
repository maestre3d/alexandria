package action

import (
	"context"
	"github.com/alexandria-oss/core/middleware"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type DeleteRequest struct {
	ID      string `json:"id"`
	Service string `json:"service"`
}

type DeleteResponse struct {
	Err error `json:"-"`
}

func MakeDeleteBlobEndpoint(svc usecase.BlobInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteRequest)
		err := svc.Delete(ctx, req.ID, req.Service)
		return DeleteResponse{Err: err}, nil
	}

	action := "delete"
	ep = middleware.WrapResiliency(ep, "blob", action)
	return middleware.WrapInstrumentation(ep, "blob", action, &middleware.WrapInstrumentParams{
		Logger:       logger,
		Duration:     duration,
		Tracer:       tracer,
		ZipkinTracer: zipkinTracer,
	})
}

var (
	_ endpoint.Failer = DeleteResponse{}
)

func (r DeleteResponse) Failed() error { return r.Err }

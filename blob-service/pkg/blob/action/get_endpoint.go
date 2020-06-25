package action

import (
	"context"
	"github.com/alexandria-oss/core/middleware"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type GetRequest struct {
	ID      string `json:"id"`
	Service string `json:"service"`
}

type GetResponse struct {
	Blob *domain.Blob `json:"blob"`
	Err  error        `json:"-"`
}

func MakeGetBlobEndpoint(svc usecase.BlobInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetRequest)
		blob, err := svc.Get(ctx, req.ID, req.Service)
		if err != nil {
			return GetResponse{
				Blob: nil,
				Err:  err,
			}, nil
		}

		return GetResponse{
			Blob: blob,
			Err:  nil,
		}, nil
	}

	// Required patterns
	action := "get"
	ep = middleware.WrapResiliency(ep, "blob", action)
	return middleware.WrapInstrumentation(ep, "blob", action, &middleware.WrapInstrumentParams{
		Logger:       logger,
		Duration:     duration,
		Tracer:       tracer,
		ZipkinTracer: zipkinTracer,
	})
}

// Compile-time failer interface assertion
var (
	_ endpoint.Failer = GetResponse{}
)

func (r GetResponse) Failed() error { return r.Err }

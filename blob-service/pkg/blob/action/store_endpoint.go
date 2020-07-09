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

type StoreRequest struct {
	RootID    string      `json:"root_id"`
	Service   string      `json:"service"`
	BlobType  string      `json:"blob_type"`
	Extension string      `json:"extension"`
	Size      string      `json:"size"`
	Content   domain.File `json:"content"`
}

type StoreResponse struct {
	Blob *domain.Blob
	Err  error `json:"-"`
}

func MakeStoreBlobEndpoint(svc usecase.BlobInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(StoreRequest)
		blob, err := svc.Store(ctx, &domain.BlobAggregate{
			RootID:    req.RootID,
			Service:   req.Service,
			BlobType:  req.BlobType,
			Extension: req.Extension,
			Size:      req.Size,
			Content:   req.Content,
		})
		if err != nil {
			return StoreResponse{
				Blob: nil,
				Err:  err,
			}, nil
		}

		return StoreResponse{
			Blob: blob,
			Err:  nil,
		}, nil
	}

	action := "store"
	ep = middleware.WrapResiliency(ep, "blob", action)
	return middleware.WrapInstrumentation(ep, "blob", action, &middleware.WrapInstrumentParams{
		Logger:       logger,
		Duration:     duration,
		Tracer:       tracer,
		ZipkinTracer: zipkinTracer,
	})
}

var (
	_ endpoint.Failer = StoreResponse{}
)

func (r StoreResponse) Failed() error { return r.Err }

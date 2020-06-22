package action

import (
	"context"
	"github.com/alexandria-oss/core/middleware"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type UpdateRequest struct {
	ID            string `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DisplayName   string `json:"display_name"`
	OwnerID       string `json:"owner_id"`
	OwnershipType string `json:"ownership_type"`
	Verified      string `json:"verified"`
	Picture       string `json:"picture"`
	Country       string `json:"country"`
}

type UpdateResponse struct {
	Author *domain.Author `json:"author"`
	Err    error          `json:"-"`
}

func MakeUpdateAuthorEndpoint(svc usecase.AuthorInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UpdateRequest)

		aggr := &domain.AuthorUpdateAggregate{
			ID: req.ID,
			RootAggregate: &domain.AuthorAggregate{
				FirstName:     req.FirstName,
				LastName:      req.LastName,
				DisplayName:   req.DisplayName,
				OwnershipType: req.OwnershipType,
				OwnerID:       req.OwnerID,
				Country:       req.Country,
			},
			Verified: req.Verified,
			Picture:  req.Picture,
		}

		author, err := svc.Update(ctx, aggr)
		if err != nil {
			return UpdateResponse{
				Author: nil,
				Err:    err,
			}, nil
		}

		return UpdateResponse{
			Author: author,
			Err:    nil,
		}, nil
	}

	// Required resiliency and instrumentation
	action := "update"
	ep = middleware.WrapResiliency(ep, "author", action)
	return middleware.WrapInstrumentation(ep, "author", action, &middleware.WrapInstrumentParams{
		logger,
		duration,
		tracer,
		zipkinTracer,
	})
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = UpdateResponse{}
)

func (r UpdateResponse) Failed() error { return r.Err }

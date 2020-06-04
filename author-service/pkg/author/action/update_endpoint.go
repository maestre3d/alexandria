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

type UpdateRequest struct {
	ID            string            `json:"id"`
	FirstName     string            `json:"first_name"`
	LastName      string            `json:"last_name"`
	DisplayName   string            `json:"display_name"`
	OwnershipType string            `json:"ownership_type"`
	OwnerID       string            `json:"owner_id"`
	Verified      string            `json:"verified"`
	Picture       string            `json:"picture"`
	Owners        map[string]string `json:"owners"`
}

type UpdateResponse struct {
	Author *domain.Author `json:"author"`
	Err    error          `json:"-"`
}

func MakeUpdateAuthorEndpoint(svc usecase.AuthorInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UpdateRequest)
		rootAggr := &domain.AuthorAggregate{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			DisplayName:   req.DisplayName,
			OwnershipType: req.OwnershipType,
			OwnerID:       req.OwnerID,
		}

		aggr := &domain.AuthorUpdateAggregate{
			ID:            req.ID,
			RootAggregate: rootAggr,
			Owners:        req.Owners,
			Verified:      req.Verified,
			Picture:       req.Picture,
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
	_ endpoint.Failer = UpdateResponse{}
)

func (r UpdateResponse) Failed() error { return r.Err }

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

type CreateRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DisplayName   string `json:"display_name"`
	OwnershipType string `json:"ownership_type"`
	OwnerID       string `json:"owner_id"`
}

type CreateResponse struct {
	Author *domain.Author `json:"author"`
	Err    error          `json:"-"`
}

func MakeCreateAuthorEndpoint(svc usecase.AuthorInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateRequest)

		createdAuthor, err := svc.Create(ctx, &domain.AuthorAggregate{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			DisplayName:   req.DisplayName,
			OwnershipType: req.OwnershipType,
			OwnerID:       req.OwnerID,
		})
		if err != nil {
			return CreateResponse{
				Author: nil,
				Err:    err,
			}, err
		}

		return CreateResponse{
			Author: createdAuthor,
			Err:    nil,
		}, nil
	}

	// Required resiliency and instrumentation
	action := "create"
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
	_ endpoint.Failer = CreateResponse{}
)

func (r CreateResponse) Failed() error { return r.Err }

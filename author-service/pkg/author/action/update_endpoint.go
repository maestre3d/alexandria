package action

import (
	"context"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type UpdateRequest struct {
	ID string `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	DisplayName string `json:"display_name"`
	BirthDate string `json:"birth_date"`
}

type UpdateResponse struct {
	Author *domain.AuthorEntity `json:"author"`
	Err string `json:"err,omitempty"`
}

func MakeUpdateAuthorEndpoint(svc service.IAuthorService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(UpdateRequest)
		author, err := svc.Update(req.ID, req.FirstName, req.LastName, req.DisplayName, req.BirthDate)
		if err != nil {
			return UpdateResponse{
				Author: nil,
				Err:    err.Error(),
			}, nil
		}

		return UpdateResponse{
			Author: author,
			Err:    "",
		}, nil
	}

	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.update"))(ep)
	ep = ratelimit.NewErroringLimiter(new(rate.Limiter))(ep)
	ep = circuitbreaker.Gobreaker(new(gobreaker.CircuitBreaker))(ep)

	return ep
}


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

type CreateRequest struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	DisplayName string `json:"display_name"`
	BirthDate string `json:"birth_date"`
}

type CreateResponse struct {
	Author *domain.AuthorEntity `json:"author"`
	Err error `json:"err,omitempty"`
}

func MakeCreateAuthorEndpoint(svc service.IAuthorService, logger log.Logger) endpoint.Endpoint {
	 ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateRequest)
		createdAuthor, err := svc.Create(req.FirstName, req.LastName, req.DisplayName, req.BirthDate)
		if err != nil {
			return CreateResponse{
				Author: createdAuthor,
				Err:    err,
			}, err
		}

		return CreateResponse{
			Author: createdAuthor,
			Err:    nil,
		}, nil
	}

	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.create"))(ep)
	ep = ratelimit.NewErroringLimiter(new(rate.Limiter))(ep)
	ep = circuitbreaker.Gobreaker(new(gobreaker.CircuitBreaker))(ep)

	return ep
}
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

type GetRequest struct {
	ID string `json:"id"`
}

type GetResponse struct {
	Author *domain.AuthorEntity `json:"author"`
	Err string `json:"err,omitempty"`
}

func MakeGetAuthorEndpoint(svc service.IAuthorService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		author, err := svc.Get(req.ID)
		if err != nil {
			return GetResponse{
				Author: nil,
				Err:    err.Error(),
			}, nil
		}

		return GetResponse{
			Author: author,
			Err:    "",
		}, nil
	}

	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.get"))(ep)
	ep = ratelimit.NewErroringLimiter(new(rate.Limiter))(ep)
	ep = circuitbreaker.Gobreaker(new(gobreaker.CircuitBreaker))(ep)

	return ep
}

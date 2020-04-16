package action

import (
	"context"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type DeleteRequest struct {
	ID string `json:"id"`
}

type DeleteResponse struct {
	Err string `json:"err,omitempty"`
}

func MakeDeleteAuthorEndpoint(svc service.IAuthorService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DeleteRequest)
		err = svc.Delete(req.ID)
		if err != nil {
			return DeleteResponse{ err.Error()}, nil
		}

		return DeleteResponse{""}, nil
	}

	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.delete"))(ep)
	ep = ratelimit.NewErroringLimiter(new(rate.Limiter))(ep)
	ep = circuitbreaker.Gobreaker(new(gobreaker.CircuitBreaker))(ep)

	return ep
}

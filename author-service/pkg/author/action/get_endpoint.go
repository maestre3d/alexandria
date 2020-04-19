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
	"time"
)

type GetRequest struct {
	ID string `json:"id"`
}

type GetResponse struct {
	Author *domain.AuthorEntity `json:"author"`
	Err    error                `json:"-"`
}

func MakeGetAuthorEndpoint(svc service.IAuthorService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		author, err := svc.Get(req.ID)
		if err != nil {
			return GetResponse{
				Author: nil,
				Err:    err,
			}, nil
		}

		return GetResponse{
			Author: author,
			Err:    nil,
		}, nil
	}

	limiter := rate.NewLimiter(rate.Every(30*time.Second), 100)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "author.get",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       0,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})

	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.get"))(ep)
	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)

	return ep
}

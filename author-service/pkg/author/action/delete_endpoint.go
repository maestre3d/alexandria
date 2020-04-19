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
	"time"
)

type DeleteRequest struct {
	ID string `json:"id"`
}

type DeleteResponse struct {
	Err error `json:"-"`
}

func MakeDeleteAuthorEndpoint(svc service.IAuthorService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DeleteRequest)
		err = svc.Delete(req.ID)
		if err != nil {
			return DeleteResponse{err}, nil
		}

		return DeleteResponse{nil}, nil
	}

	limiter := rate.NewLimiter(rate.Every(30*time.Second), 100)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "author.delete",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       0,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})

	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.delete"))(ep)
	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)

	return ep
}

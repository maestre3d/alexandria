package action

import (
	"context"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

type ListRequest struct {
	PageToken    string            `json:"page_token"`
	PageSize     string            `json:"page_size"`
	FilterParams util.FilterParams `json:"filter_params"`
}

type ListResponse struct {
	Authors       []*domain.AuthorEntity `json:"authors"`
	NextPageToken string                 `json:"next_page_token"`
	Err           error                  `json:"-"`
}

func MakeListAuthorEndpoint(svc service.IAuthorService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ListRequest)
		authors, nextToken, err := svc.List(req.PageToken, req.PageSize, req.FilterParams)
		if err != nil {
			return ListResponse{
				Authors:       nil,
				NextPageToken: "",
				Err:           err,
			}, nil
		}

		return ListResponse{
			Authors:       authors,
			NextPageToken: nextToken,
			Err:           nil,
		}, nil
	}

	limiter := rate.NewLimiter(rate.Every(30*time.Second), 100)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "author.list",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       0,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})

	ep = shared.LoggingMiddleware(log.With(logger, "method", "author.list"))(ep)
	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)

	return ep
}

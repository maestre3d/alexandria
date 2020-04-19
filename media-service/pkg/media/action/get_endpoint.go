package action

import (
	"context"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
	"github.com/maestre3d/alexandria/media-service/pkg/shared"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
	"time"
)

type GetRequest struct {
	ID string `json:"id"`
}

type GetResponse struct {
	Media *domain.MediaEntity `json:"media"`
	Err   error               `json:"-"`
}

func MakeGetMediaEndpoint(svc service.IMediaService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		media, err := svc.Get(req.ID)
		if err != nil {
			return GetResponse{
				Media: nil,
				Err:   err,
			}, nil
		}

		return GetResponse{
			Media: media,
			Err:   nil,
		}, nil
	}

	limiter := rate.NewLimiter(rate.Every(30*time.Second), 100)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "media.get",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       0,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})

	ep = shared.LoggingMiddleware(log.With(logger, "method", "media.get"))(ep)
	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)

	return ep
}

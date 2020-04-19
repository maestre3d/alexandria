package action

import (
	"context"

	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
	"github.com/maestre3d/alexandria/media-service/pkg/shared"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

type CreateRequest struct {
	Title       string `json:"title"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	UserID      string `json:"user_id"`
	AuthorID    string `json:"author_id"`
	PublishDate string `json:"publish_date"`
	MediaType   string `json:"media_type"`
}

type CreateResponse struct {
	Media *domain.MediaEntity `json:"media"`
	Err   error               `json:"-"`
}

func MakeCreateMediaEndpoint(svc service.IMediaService, logger log.Logger) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateRequest)
		createdMedia, err := svc.Create(req.Title, req.DisplayName, req.Description, req.UserID, req.AuthorID, req.PublishDate, req.MediaType)
		if err != nil {
			return CreateResponse{
				Media: nil,
				Err:   err,
			}, nil
		}

		return CreateResponse{
			Media: createdMedia,
			Err:   nil,
		}, nil
	}

	limiter := rate.NewLimiter(rate.Every(30*time.Second), 100)
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          "media.create",
		MaxRequests:   100,
		Interval:      0,
		Timeout:       0,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})

	ep = shared.LoggingMiddleware(log.With(logger, "method", "media.create"))(ep)
	ep = ratelimit.NewErroringLimiter(limiter)(ep)
	ep = circuitbreaker.Gobreaker(cb)(ep)

	return ep
}

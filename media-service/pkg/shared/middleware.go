package shared

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			err := logger.Log("msg", "calling endpoint")
			if err != nil {
				return nil, err
			}
			defer func() {
				err = logger.Log("msg", "called endpoint")
			}()
			return next(ctx, request)
		}
	}
}

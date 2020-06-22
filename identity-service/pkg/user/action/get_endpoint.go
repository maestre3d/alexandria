package action

import (
	"context"
	"github.com/alexandria-oss/core/middleware"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type GetRequest struct {
	ID string `json:"id"`
}

type GetResponse struct {
	User *domain.User `json:"user"`
	Err    error          `json:"-"`
}

func MakeGetUserEndpoint(svc usecase.UserInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetRequest)
		user, err := svc.Get(ctx, req.ID)
		if err != nil {
			return GetResponse{
				User: nil,
				Err:    err,
			}, nil
		}

		return GetResponse{
			User: user,
			Err:    nil,
		}, nil
	}

	// Required resiliency and instrumentation
	action := "get"
	ep = middleware.WrapResiliency(ep, "user", action)
	return middleware.WrapInstrumentation(ep, "user", action, &middleware.WrapInstrumentParams{
		logger,
		duration,
		tracer,
		zipkinTracer,
	})
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = GetResponse{}
)

func (r GetResponse) Failed() error { return r.Err }

package user

import (
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/middleware"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// WrapUserInstrumentation Inject middleware (metrics and logging) to bounded context's edge
// using chain of responsibility/middleware pattern and HOC-like pattern wrapping style
func WrapUserInstrumentation(userUseCase usecase.UserInteractor, logger log.Logger) usecase.UserInteractor {
	// Inject prometheus and OpenCensus metrics
	// Inject logger
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "identity_service",
		Name:        "request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "identity_service",
		Name:        "request_latency",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, fieldKeys)

	var svc usecase.UserInteractor
	svc = userUseCase
	svc = middleware.LoggingUserMiddleware{Logger: logger, Next: svc}
	svc = middleware.MetricUserMiddleware{RequestCount: requestCount, RequestLatency: requestLatency, Next: svc}

	return svc
}

// WrapUserSAGAInstrumentation Inject middleware (metrics and logging) to bounded context's edge
// using chain of responsibility/middleware pattern and HOC-like pattern wrapping style
func WrapUserSAGAInstrumentation(userUseCase usecase.UserSAGAInteractor, logger log.Logger) usecase.UserSAGAInteractor {
	// Inject prometheus and OpenCensus metrics
	// Inject logger
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "identity_service",
		Name:        "saga_request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "identity_service",
		Name:        "saga_request_latency",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, fieldKeys)

	var svc usecase.UserSAGAInteractor
	svc = userUseCase
	svc = middleware.LoggingUserSAGAMiddleware{Logger: logger, Next: svc}
	svc = middleware.MetricUserSAGAMiddleware{RequestCount: requestCount, RequestLatency: requestLatency, Next: svc}

	return svc
}

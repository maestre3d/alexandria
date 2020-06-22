package author

import (
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/maestre3d/alexandria/author-service/pkg/author/middleware"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// WrapAuthorInstrumentation
// Inject middleware (metrics and logging) to bounded context's edge
// using chain of responsibility/middleware pattern and HOC-like pattern wrapping style
func WrapAuthorInstrumentation(authorUseCase usecase.AuthorInteractor, logger log.Logger) usecase.AuthorInteractor {
	// Inject prometheus and OpenCensus metrics
	// Inject logger
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "author_service",
		Name:        "request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "author_service",
		Name:        "request_latency",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, fieldKeys)

	var svc usecase.AuthorInteractor
	svc = authorUseCase
	svc = middleware.LoggingAuthorMiddleware{Logger: logger, Next: svc}
	svc = middleware.MetricAuthorMiddleware{RequestCount: requestCount, RequestLatency: requestLatency, Next: svc}

	return svc
}

// WrapAuthorSAGAInstrumentation
// Inject middleware (metrics and logging) to bounded context's edge
// using chain of responsibility/middleware pattern and HOC-like pattern wrapping style
func WrapAuthorSAGAInstrumentation(authorUseCase usecase.AuthorSAGAInteractor, logger log.Logger) usecase.AuthorSAGAInteractor {
	// Inject prometheus and OpenCensus metrics
	// Inject logger
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "author_service",
		Name:        "saga_request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "author_service",
		Name:        "saga_request_latency",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, fieldKeys)

	var svc usecase.AuthorSAGAInteractor
	svc = authorUseCase
	svc = middleware.LoggingAuthorSAGAMiddleware{Logger: logger, Next: svc}
	svc = middleware.MetricAuthorSAGAMiddleware{RequestCount: requestCount, RequestLatency: requestLatency, Next: svc}

	return svc
}

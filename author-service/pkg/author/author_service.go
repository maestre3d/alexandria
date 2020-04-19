package author

import (
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/maestre3d/alexandria/author-service/pkg/author/middleware"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func NewAuthorService(authorUseCase service.IAuthorService, logger log.Logger) service.IAuthorService {
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

	var svc service.IAuthorService
	svc = authorUseCase
	svc = middleware.LoggingAuthorMiddleware{logger, svc}
	svc = middleware.InstrumentingAuthorMiddleware{requestCount, requestLatency, svc}

	return svc
}

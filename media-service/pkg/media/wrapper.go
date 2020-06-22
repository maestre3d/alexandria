package media

import (
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/maestre3d/alexandria/media-service/pkg/media/middleware"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// WrapMediaInstrumentation Inject middleware (metrics and logging) to bounded context's edge
// using chain of responsibility/middleware pattern and HOC-like pattern wrapping style
func WrapMediaInstrumentation(MediaUseCase usecase.MediaInteractor, logger log.Logger) usecase.MediaInteractor {
	// Inject Prometheus metrics
	// Inject logger
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "media_service",
		Name:        "request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "media_service",
		Name:        "request_latency",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, fieldKeys)
	requestGauge := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace:   "alexandria",
		Subsystem:   "media_service",
		Name:        "request_memory_usage",
		Help:        "request memory consumption",
		ConstLabels: nil,
	}, fieldKeys)

	var svc usecase.MediaInteractor
	svc = MediaUseCase
	svc = middleware.LoggingMediaMiddleware{Logger: logger, Next: svc}
	svc = middleware.MetricMediaMiddleware{RequestCount: requestCount, RequestLatency: requestLatency, RequestGauge: requestGauge, Next: svc}

	return svc
}

// WrapMediaSAGAInstrumentation Inject middleware (metrics and logging) to bounded context's edge
// using chain of responsibility/middleware pattern and HOC-like pattern wrapping style
func WrapMediaSAGAInstrumentation(MediaUseCase usecase.MediaSAGAInteractor, logger log.Logger) usecase.MediaSAGAInteractor {
	// Inject Prometheus metrics
	// Inject logger
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "media_service",
		Name:        "saga_request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "media_service",
		Name:        "saga_request_latency",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, fieldKeys)
	requestGauge := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace:   "alexandria",
		Subsystem:   "media_service",
		Name:        "saga_request_memory_usage",
		Help:        "request memory consumption",
		ConstLabels: nil,
	}, fieldKeys)

	var svc usecase.MediaSAGAInteractor
	svc = MediaUseCase
	svc = middleware.LoggingMediaSAGAMiddleware{Logger: logger, Next: svc}
	svc = middleware.MetricMediaSAGAMiddleware{RequestCount: requestCount, RequestLatency: requestLatency, RequestGauge: requestGauge, Next: svc}

	return svc
}

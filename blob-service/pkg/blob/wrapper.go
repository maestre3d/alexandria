package blob

import (
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/middleware"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func WrapBlobInstrumentation(blobUseCase usecase.BlobInteractor, logger log.Logger) usecase.BlobInteractor {
	// Inject Prometheus metrics
	// Inject logger
	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "blob_service",
		Name:        "request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace:   "alexandria",
		Subsystem:   "blob_service",
		Name:        "request_latency",
		Help:        "total duration of requests in microseconds",
		ConstLabels: nil,
		Objectives:  nil,
		MaxAge:      0,
		AgeBuckets:  0,
		BufCap:      0,
	}, fieldKeys)

	var svc usecase.BlobInteractor
	svc = blobUseCase
	svc = middleware.LoggingBlobMiddleware{
		Logger: logger,
		Next:   svc,
	}
	svc = middleware.MetricBlobMiddleware{
		RequestCount:   requestCount,
		RequestLatency: requestLatency,
		Next:           svc,
	}

	return svc
}

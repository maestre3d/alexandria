package media

import (
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/maestre3d/alexandria/media-service/pkg/media/middleware"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func NewMediaService(mediaUseCase service.IMediaService, logger log.Logger) service.IMediaService {
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

	var svc service.IMediaService
	svc = mediaUseCase
	svc = middleware.LoggingMediaMiddleware{logger, svc}
	svc = middleware.InstrumentingMediaMiddleware{requestCount, requestLatency, svc}

	return svc
}

package wrapper

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/category-service/pkg/middleware"
	"github.com/maestre3d/alexandria/category-service/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/ratelimit"
)

// HOC-like function to attach required observability (tracing, logging & metrics) and
// resiliency patterns to category's use case layer using chain-of-responsibility pattern
func WrapCategoryMiddleware(svcUnwrap service.Category, logger log.Logger) service.Category {
	var svc service.Category
	svc = svcUnwrap
	svc = middleware.CategoryLog{Logger: logger, Next: svc}
	svc = injectMetrics(svc, logger)
	svc = middleware.CategoryResiliency{
		Logger:      logger,
		RateLimiter: ratelimit.New(100),
		Next:        svc,
	}
	return svc
}

func injectMetrics(svc service.Category, logger log.Logger) service.Category {
	labels := []string{"method", "error"}
	requestLatency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "alexandria",
		Subsystem:   "category_service",
		Name:        "request_latency",
		Help:        "total duration of request in microseconds",
		ConstLabels: nil,
		Buckets:     prometheus.DefBuckets,
	}, labels)
	requestCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "category_service",
		Name:        "request_count",
		Help:        "number of request received",
		ConstLabels: nil,
	}, labels)
	requestErrorCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "alexandria",
		Subsystem:   "category_service",
		Name:        "request_error_count",
		Help:        "number of errors from request received",
		ConstLabels: nil,
	}, []string{"method"})
	categoryTotal := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "alexandria",
		Subsystem:   "category_service",
		Name:        "categories_total",
		Help:        "total categories",
		ConstLabels: nil,
	})

	// DO not panic or return errors, metrics are optional
	err := prometheus.Register(requestLatency)
	if err != nil {
		_ = level.Warn(logger).Log(
			"msg", "prometheus has failed to register request_latency histogram",
			"metric_name", "request_latency",
			"kind", "histogram",
			"err", err,
		)
	}

	err = prometheus.Register(requestErrorCount)
	if err != nil {
		_ = level.Warn(logger).Log(
			"msg", "prometheus has failed to register request_error_count counter",
			"metric_name", "request_error_count",
			"kind", "counter",
			"err", err,
		)
	}

	err = prometheus.Register(categoryTotal)
	if err != nil {
		_ = level.Warn(logger).Log(
			"msg", "prometheus has failed to register categories_total gauge",
			"metric_name", "categories_total",
			"kind", "gauge",
			"err", err,
		)
	}

	err = prometheus.Register(requestCount)
	if err != nil {
		_ = level.Warn(logger).Log(
			"msg", "prometheus has failed to register request_count counter, avoiding category_metric injection",
			"metric_name", "request_count",
			"kind", "counter",
			"err", err,
		)
		return svc
	}

	return middleware.CategoryMetric{
		ReqCounter:      requestCount,
		ReqHistogram:    requestLatency,
		CategoriesTotal: categoryTotal,
		Next:            svc,
	}
}

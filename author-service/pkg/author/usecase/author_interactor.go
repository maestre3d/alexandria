package usecase

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/middleware"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type AuthorInteractor interface {
	Create(ctx context.Context, aggregate *domain.AuthorAggregate) (*domain.Author, error)
	List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) ([]*domain.Author, string, error)
	Get(ctx context.Context, id string) (*domain.Author, error)
	Update(ctx context.Context, id, status string, aggregate *domain.AuthorAggregate) (*domain.Author, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

// WrapAuthorInstrumentation Inject middleware (metrics and logging) to bounded context's edge
// using chain of responsibility/middleware pattern and HOC-like pattern wrapping style
func WrapAuthorInstrumentation(authorUseCase AuthorInteractor, logger log.Logger) AuthorInteractor {
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

	var svc AuthorInteractor
	svc = authorUseCase
	svc = middleware.LoggingAuthorMiddleware{logger, svc}
	svc = middleware.MetricAuthorMiddleware{requestCount, requestLatency, svc}

	return svc
}

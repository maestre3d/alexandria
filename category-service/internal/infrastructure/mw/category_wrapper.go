package mw

import (
	"github.com/alexandria-oss/core/config"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"sync"
)

// HOC-like function to inject observability to the event bus implementation
func WrapCategoryEventObservability(svcUnwrap domain.CategoryEventBus, logger log.Logger) domain.CategoryEventBus {
	var svc domain.CategoryEventBus
	svc = svcUnwrap
	svc = CategoryEventLog{
		Logger: logger,
		Next:   svc,
	}
	// Add metrics with prometheus + OpenCensus if needed
	svc = CategoryEventTracing{Next: svc}

	return svc
}

func WrapCategoryRepoTools(svcUnwrap domain.CategoryRepository, redisPool *redis.Client, cfg *config.Kernel) domain.CategoryRepository {
	var svc domain.CategoryRepository
	svc = svcUnwrap
	svc = CategoryRepositoryCache{
		Pool: redisPool,
		Cfg:  cfg,
		Next: svc,
		Mu:   new(sync.RWMutex),
	}
	svc = CategoryRepositoryMetric{Next: svc}
	svc = CategoryRepositoryTracing{Next: svc}

	return svc
}

// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/category-service/internal/infrastructure/cassandra"
	"github.com/maestre3d/alexandria/category-service/internal/infrastructure/mw"
	"github.com/maestre3d/alexandria/category-service/internal/interactor"
)

// Injectors from wire.go:

func InjectCategoryUseCase() (*interactor.CategoryUseCase, func(), error) {
	logLogger := logger.NewZapLogger()
	context := provideContext()
	kernel, err := config.NewKernel(context)
	if err != nil {
		return nil, nil, err
	}
	clusterConfig := cassandra.NewCassandraPool(kernel)
	categoryRepositoryCassandra := infrastructure.NewCategoryRepositoryCassandra(clusterConfig)
	client, cleanup, err := persistence.NewRedisPool(kernel)
	if err != nil {
		return nil, nil, err
	}
	categoryRepository := provideCategoryRepository(categoryRepositoryCassandra, client, kernel)
	categoryEventKafka := infrastructure.NewCategoryEventKafka(kernel)
	categoryEventBus := provideCategoryEventBus(categoryEventKafka, logLogger)
	categoryUseCase := interactor.NewCategoryUseCase(logLogger, categoryRepository, categoryEventBus)
	return categoryUseCase, func() {
		cleanup()
	}, nil
}

func InjectCategoryRootUseCase() (*interactor.CategoryRootUseCase, func(), error) {
	logLogger := logger.NewZapLogger()
	context := provideContext()
	kernel, err := config.NewKernel(context)
	if err != nil {
		return nil, nil, err
	}
	clusterConfig := cassandra.NewCassandraPool(kernel)
	categoryRootCassandraRepository := infrastructure.NewCategoryRootCassandraRepository(clusterConfig)
	categoryRepositoryCassandra := infrastructure.NewCategoryRepositoryCassandra(clusterConfig)
	client, cleanup, err := persistence.NewRedisPool(kernel)
	if err != nil {
		return nil, nil, err
	}
	categoryRepository := provideCategoryRepository(categoryRepositoryCassandra, client, kernel)
	categoryRootUseCase := interactor.NewCategoryRootUseCase(logLogger, categoryRootCassandraRepository, categoryRepository)
	return categoryRootUseCase, func() {
		cleanup()
	}, nil
}

// wire.go:

var ctx = context.Background()

var dataSet = wire.NewSet(
	provideContext, config.NewKernel, persistence.NewRedisPool, cassandra.NewCassandraPool, infrastructure.NewCategoryRepositoryCassandra, provideCategoryRepository,
)

func SetContext(ctxRoot context.Context) {
	ctx = ctxRoot
}

func provideContext() context.Context {
	return ctx
}

func provideCategoryEventBus(eventBus *infrastructure.CategoryEventKafka, loggerImp log.Logger) domain.CategoryEventBus {
	return mw.WrapCategoryEventObservability(eventBus, loggerImp)
}

func provideCategoryRepository(repo *infrastructure.CategoryRepositoryCassandra, redis2 *redis.Client, cfg *config.Kernel) domain.CategoryRepository {
	return mw.WrapCategoryRepoTools(repo, redis2, cfg)
}

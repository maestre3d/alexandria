// +build wireinject

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

var ctx = context.Background()

var dataSet = wire.NewSet(
	provideContext,
	config.NewKernel,
	persistence.NewRedisPool,
	cassandra.NewCassandraPool,
	infrastructure.NewCategoryRepositoryCassandra,
	provideCategoryRepository,
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

func provideCategoryRepository(repo *infrastructure.CategoryRepositoryCassandra, redis *redis.Client, cfg *config.Kernel) domain.CategoryRepository {
	return mw.WrapCategoryRepoTools(repo, redis, cfg)
}

func InjectCategoryUseCase() (*interactor.CategoryUseCase, func(), error) {
	wire.Build(
		dataSet,
		logger.NewZapLogger,
		infrastructure.NewCategoryEventKafka,
		provideCategoryEventBus,
		interactor.NewCategoryUseCase,
	)

	return &interactor.CategoryUseCase{}, nil, nil
}

func InjectCategoryRootUseCase() (*interactor.CategoryRootUseCase, func(), error) {
	wire.Build(
		dataSet,
		logger.NewZapLogger,
		wire.Bind(new(domain.CategoryRootRepository), new(*infrastructure.CategoryRootCassandraRepository)),
		infrastructure.NewCategoryRootCassandraRepository,
		interactor.NewCategoryRootUseCase,
	)

	return &interactor.CategoryRootUseCase{}, nil, nil
}

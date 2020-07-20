// +build wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/category-service/internal/interactor"
)

var ctx = context.Background()

var dataSet = wire.NewSet(
	provideContext,
	config.NewKernel,
	persistence.NewRedisPool,
	infrastructure.NewCassandraPool,
	wire.Bind(new(domain.CategoryRepository), new(*infrastructure.CategoryRepositoryCassandra)),
	infrastructure.NewCategoryRepositoryCassandra,
)

func SetContext(ctxRoot context.Context) {
	ctx = ctxRoot
}

func provideContext() context.Context {
	return ctx
}

func InjectCategoryUseCase() (*interactor.CategoryUseCase, func(), error) {
	wire.Build(
		dataSet,
		logger.NewZapLogger,
		wire.Bind(new(domain.CategoryEventBus), new(*infrastructure.CategoryEventKafka)),
		infrastructure.NewCategoryEventKafka,
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

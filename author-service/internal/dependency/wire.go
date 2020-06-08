// +build wireinject

package dependency

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/maestre3d/alexandria/author-service/internal/infrastructure"

	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/internal/interactor"
)

var configSet = wire.NewSet(
	provideContext,
	config.NewKernel,
)

var dBMSPoolSet = wire.NewSet(
	configSet,
	persistence.NewPostgresPool,
)

var authorDBMSRepositorySet = wire.NewSet(
	dBMSPoolSet,
	logger.NewZapLogger,
	persistence.NewRedisPool,
	wire.Bind(new(domain.IAuthorRepository), new(*infrastructure.AuthorPostgresRepository)),
	infrastructure.NewAuthorPostgresRepository,
)

func provideContext() context.Context {
	return context.Background()
}

func InjectAuthorUseCase() (*interactor.AuthorUseCase, func(), error) {
	wire.Build(
		authorDBMSRepositorySet,
		wire.Bind(new(domain.IAuthorEventBus), new(*infrastructure.AuthorKafkaEventBus)),
		infrastructure.NewAuthorAWSEventBus,
		interactor.NewAuthorUseCase,
	)

	return &interactor.AuthorUseCase{}, nil, nil
}

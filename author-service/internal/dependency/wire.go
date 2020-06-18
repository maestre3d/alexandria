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

var Ctx context.Context = context.Background()

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
	wire.Bind(new(domain.AuthorRepository), new(*infrastructure.AuthorPostgresRepository)),
	infrastructure.NewAuthorPostgresRepository,
)

func provideContext() context.Context {
	return Ctx
}

func InjectAuthorUseCase() (*interactor.AuthorUseCase, func(), error) {
	wire.Build(
		authorDBMSRepositorySet,
		wire.Bind(new(domain.AuthorEventBus), new(*infrastructure.AuthorKafkaEventBus)),
		infrastructure.NewAuthorKafkaEventBus,
		interactor.NewAuthorUseCase,
	)

	return &interactor.AuthorUseCase{}, nil, nil
}

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

var dataSet = wire.NewSet(
	provideContext,
	config.NewKernel,
	persistence.NewPostgresPool,
	persistence.NewRedisPool,
	logger.NewZapLogger,
	wire.Bind(new(domain.AuthorRepository), new(*infrastructure.AuthorPQRepository)),
	infrastructure.NewAuthorPQRepository,
)

var eventSet = wire.NewSet(
	wire.Bind(new(domain.AuthorEventBus), new(*infrastructure.AuthorKafkaEventBus)),
	infrastructure.NewAuthorKafkaEventBus,
)

func provideContext() context.Context {
	return Ctx
}

func InjectAuthorUseCase() (*interactor.Author, func(), error) {
	wire.Build(
		dataSet,
		eventSet,
		interactor.NewAuthor,
	)

	return &interactor.Author{}, nil, nil
}

func InjectAuthorSAGAUseCase() (*interactor.AuthorSAGA, func(), error) {
	wire.Build(
		dataSet,
		eventSet,
		wire.Bind(new(domain.AuthorSAGAEventBus), new(*infrastructure.AuthorSAGAKafkaEventBus)),
		infrastructure.NewAuthorSAGAKafkaEventBus,
		interactor.NewAuthorSAGA,
	)

	return &interactor.AuthorSAGA{}, nil, nil
}

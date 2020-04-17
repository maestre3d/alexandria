// +build wireinject

package dependency

import (
	"context"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/author/infrastructure"
	"github.com/maestre3d/alexandria/author-service/internal/author/interactor"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/logging"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/persistence"
)

var LoggerSet = wire.NewSet(
	wire.Bind(new(util.ILogger) , new(*logging.ZapLogger)),
	logging.NewZapLogger,
)

var DBMSPoolSet = wire.NewSet(
	ProvideContext,
	LoggerSet,
	persistence.NewPostgresPool,
)

var AuthorDBMSRepositorySet = wire.NewSet(
	DBMSPoolSet,
	wire.Bind(new(domain.IAuthorRepository), new(*infrastructure.AuthorDBMSRepository)),
	infrastructure.NewAuthorDBMSRepository,
)

var AuthorServiceSet = wire.NewSet(
	AuthorDBMSRepositorySet,
	interactor.NewAuthorUseCase,
)

func ProvideContext() context.Context {
	return context.Background()
}

func InjectAuthorUseCase() (*interactor.AuthorUseCase, func(), error) {
	wire.Build(
		AuthorServiceSet,
	)

	return &interactor.AuthorUseCase{}, nil, nil
}
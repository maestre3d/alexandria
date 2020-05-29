//+build wireinject

package dependency

import (
	"context"
	"database/sql"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/identity-service/internal/interactor"
)

var cfgSet = wire.NewSet(
	provideContext,
	config.NewKernel,
)

var logSet = wire.NewSet(
	logger.NewZapLogger,
)

var persistenceSet = wire.NewSet(
	cfgSet,
	persistence.NewPostgresPool,
)

var userRepository = wire.NewSet(
	persistenceSet,
	logSet,
	provideUserPostgresRepository,
)

var userUseCase = wire.NewSet(
	userRepository,
	interactor.NewUserUseCase,
)

func provideContext() context.Context {
	return context.Background()
}

func provideUserPostgresRepository(ctx context.Context, logger log.Logger, db *sql.DB, cfg *config.Kernel) (domain.UserRepository, func()) {
	repo := infrastructure.NewUserPostgresRepository(logger, db)
	mem, cleanup, _ := persistence.NewRedisPool(cfg)

	repo.SetInMem(mem)

	return repo, cleanup
}

func InjectUserUseCase() (*interactor.UserUseCase, func(), error) {
	wire.Build(userUseCase)
	return &interactor.UserUseCase{}, nil, nil
}

func InjectIdentityUseCase() (*interactor.IdentityUseCase, func(), error) {
	wire.Build(
		cfgSet,
		logSet,
		wire.Bind(new(domain.ProviderAdapter), new(*infrastructure.IdentityCognitoAdapter)),
		infrastructure.NewProviderCognitoAdapter,
		interactor.NewIdentityUseCase,
	)

	return &interactor.IdentityUseCase{}, nil, nil
}

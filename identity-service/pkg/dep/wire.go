// +build wireinject

package dep

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/transport/proxy"
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/identity-service/internal/dependency"
	"github.com/maestre3d/alexandria/identity-service/pkg/service"
	"github.com/maestre3d/alexandria/identity-service/pkg/transport/bind"
	"github.com/maestre3d/alexandria/identity-service/pkg/user"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
)

var Ctx context.Context = context.Background()

var userSAGAInteractorSet = wire.NewSet(
	logger.NewZapLogger,
	provideUserSAGAInteractor,
)
var eventProxySet = wire.NewSet(
	userSAGAInteractorSet,
	provideContext,
	config.NewKernel,
	bind.NewUserEventConsumer,
	provideEventConsumers,
	proxy.NewEvent,
)

func provideContext() context.Context {
	return Ctx
}

func provideUserSAGAInteractor(logger log.Logger) (usecase.UserSAGAInteractor, error) {
	dependency.Ctx = Ctx
	userUseCase, err := dependency.InjectUserSAGAUseCase()

	userService := user.WrapUserSAGAInstrumentation(userUseCase, logger)

	return userService, err
}

func provideEventConsumers(userConsumer *bind.UserEventConsumer) []proxy.Consumer {
	consumers := make([]proxy.Consumer, 0)
	consumers = append(consumers, userConsumer)
	return consumers
}

func InjectTransportService() (*service.Transport, func(), error) {
	wire.Build(eventProxySet, service.NewTransport)

	return &service.Transport{}, nil, nil
}

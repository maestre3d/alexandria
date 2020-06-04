// +build wireinject

package dep

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/dependency"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/bind"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/proxy"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/tracer"
)

var authorInteractorSet = wire.NewSet(
	logger.NewZapLogger,
	provideAuthorInteractor,
)

var httpProxySet = wire.NewSet(
	authorInteractorSet,
	provideContext,
	config.NewKernel,
	tracer.NewZipkin,
	tracer.WrapOpenTracing,
	bind.NewAuthorHTTP,
	provideHTTPHandlers,
	shared.NewHTTPServer,
	proxy.NewHTTP,
)

var rpcProxySet = wire.NewSet(
	bind.NewAuthorRPC,
	provideRPCServers,
	proxy.NewRPC,
)

func provideContext() context.Context {
	return context.Background()
}

func provideAuthorInteractor(logger log.Logger) (usecase.AuthorInteractor, func(), error) {
	authorUseCase, cleanup, err := dependency.InjectAuthorUseCase()

	authorService := author.WrapAuthorInstrumentation(authorUseCase, logger)

	return authorService, cleanup, err
}

// Bind/Map used http handlers
func provideHTTPHandlers(authorHandler *bind.AuthorHandler) []proxy.Handler {
	handlers := make([]proxy.Handler, 0)
	handlers = append(handlers, authorHandler)
	return handlers
}

// Bind/Map used rpc actions
func provideRPCServers(authorHandler pb.AuthorServer) *proxy.Servers {
	return &proxy.Servers{authorHandler}
}

func InjectTransportService() (*service.Transport, func(), error) {
	wire.Build(httpProxySet, rpcProxySet, service.NewTransport)

	return &service.Transport{}, nil, nil
}

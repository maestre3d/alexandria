// +build wireinject

package di

import (
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/dependency"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/transport"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/handler"
	"os"
)

var authorServiceSet = wire.NewSet(
	ProvideLogger,
	ProvideAuthorService,
)

var proxyHandlersSet = wire.NewSet(
	authorServiceSet,
	handler.NewAuthorHandler,
	ProvideProxyHandlers,
)

var httpProxySet = wire.NewSet(
	proxyHandlersSet,
	shared.NewHTTPServer,
	transport.NewHTTPTransportProxy,
)

func ProvideLogger() log.Logger {
	return log.NewLogfmtLogger(os.Stderr)
}

func ProvideAuthorService(logger log.Logger) (service.IAuthorService, func(), error) {
	authorUsecase, cleanup, err := dependency.InjectAuthorUseCase()

	authorService := author.NewAuthorService(authorUsecase, logger)

	return authorService, cleanup, err
}

func ProvideProxyHandlers(authorHandler *handler.AuthorHandler) *transport.ProxyHandlers {
	return &transport.ProxyHandlers{authorHandler}
}

func InjectHTTPProxy() (*transport.HTTPTransportProxy, func(), error) {
	wire.Build(
		httpProxySet,
	)

	return &transport.HTTPTransportProxy{}, nil, nil
}
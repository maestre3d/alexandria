// +build wireinject

package dep

import (
	"context"
	"github.com/go-kit/kit/log"
	logzap "github.com/go-kit/kit/log/zap"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/dependency"
	"github.com/maestre3d/alexandria/author-service/internal/infrastructure/config"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/transport"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/bind"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/proxy"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/tracer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var authorServiceSet = wire.NewSet(
	provideLogger,
	provideAuthorService,
)

var proxyHandlersSet = wire.NewSet(
	authorServiceSet,
	config.NewKernelConfig,
	tracer.NewOpenTracer,
	tracer.NewZipkinTracer,
	bind.NewAuthorHandler,
	provideProxyHandlers,
)

var httpProxySet = wire.NewSet(
	proxyHandlersSet,
	provideContext,
	shared.NewHTTPServer,
	proxy.NewHTTPTransportProxy,
)

var rpcProxyHandlersSet = wire.NewSet(
	bind.NewAuthorRPCServer,
	provideRPCProxyHandlers,
)

var rpcProxySet = wire.NewSet(
	rpcProxyHandlersSet,
	proxy.NewRPCTransportProxy,
)

func provideContext() context.Context {
	return context.Background()
}

func provideLogger() log.Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	level := zapcore.Level(8)

	return logzap.NewZapSugarLogger(zapLogger, level)
}

func provideAuthorService(logger log.Logger) (usecase.IAuthorService, func(), error) {
	authorUseCase, cleanup, err := dependency.InjectAuthorUseCase()

	authorService := author.NewAuthorService(authorUseCase, logger)

	return authorService, cleanup, err
}

func provideProxyHandlers(authorHandler *bind.AuthorHandler) *proxy.ProxyHandlers {
	return &proxy.ProxyHandlers{authorHandler}
}

func provideRPCProxyHandlers(authorHandler pb.AuthorServer) *proxy.RPCProxyHandlers {
	return &proxy.RPCProxyHandlers{authorHandler}
}

func InjectTransportService() (*transport.TransportService, func(), error) {
	wire.Build(httpProxySet, rpcProxySet, transport.NewTransportService)

	return &transport.TransportService{}, nil, nil
}

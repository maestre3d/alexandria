// +build wireinject

package di

import (
	"context"
	"github.com/go-kit/kit/log"
	logzap "github.com/go-kit/kit/log/zap"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/config"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/dependency"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/transport"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/handler"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/pb"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/proxy"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/tracer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var authorServiceSet = wire.NewSet(
	ProvideLogger,
	ProvideAuthorService,
)

var proxyHandlersSet = wire.NewSet(
	authorServiceSet,
	config.NewKernelConfig,
	tracer.NewOpenTracer,
	tracer.NewZipkinTracer,
	handler.NewAuthorHandler,
	ProvideProxyHandlers,
)

var httpProxySet = wire.NewSet(
	proxyHandlersSet,
	ProvideContext,
	shared.NewHTTPServer,
	proxy.NewHTTPTransportProxy,
)

var rpcProxyHandlersSet = wire.NewSet(
	handler.NewAuthorRPCServer,
	ProvideRPCProxyHandlers,
)

var rpcProxySet = wire.NewSet(
	rpcProxyHandlersSet,
	proxy.NewRPCTransportProxy,
)

func ProvideContext() context.Context {
	return context.Background()
}

func ProvideLogger() log.Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	level := zapcore.Level(8)

	return logzap.NewZapSugarLogger(zapLogger, level)
}

func ProvideAuthorService(logger log.Logger) (service.IAuthorService, func(), error) {
	authorUseCase, cleanup, err := dependency.InjectAuthorUseCase()

	authorService := author.NewAuthorService(authorUseCase, logger)

	return authorService, cleanup, err
}

func ProvideProxyHandlers(authorHandler *handler.AuthorHandler) *proxy.ProxyHandlers {
	return &proxy.ProxyHandlers{authorHandler}
}

func ProvideRPCProxyHandlers(authorHandler pb.AuthorServer) *proxy.RPCProxyHandlers {
	return &proxy.RPCProxyHandlers{authorHandler}
}

func InjectTransportService() (*transport.TransportService, func(), error) {
	wire.Build(httpProxySet, rpcProxySet, transport.NewTransportService)

	return &transport.TransportService{}, nil, nil
}

// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dep

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/go-kit/kit/log"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/category-service/internal/dependency"
	"github.com/maestre3d/alexandria/category-service/pkg/middleware"
	"github.com/maestre3d/alexandria/category-service/pkg/service"
	"github.com/maestre3d/alexandria/category-service/pkg/transport"
	"github.com/maestre3d/alexandria/category-service/pkg/transport/handler"
)

// Injectors from wire.go:

func InjectTransportProxy() (*transport.Proxy, func(), error) {
	context := provideContext()
	kernel, err := config.NewKernel(context)
	if err != nil {
		return nil, nil, err
	}
	logLogger := logger.NewZapLogger()
	category, cleanup, err := provideCategoryService(context, logLogger)
	if err != nil {
		return nil, nil, err
	}
	categoryHTTP := handler.NewCategoryHTTP(category)
	v := provideHandlers(categoryHTTP)
	httpServer := transport.NewHTTPServer(kernel, logLogger, v...)
	proxy := transport.NewProxy(httpServer)
	return proxy, func() {
		cleanup()
	}, nil
}

// wire.go:

var ctx = context.Background()

var httpCategorySet = wire.NewSet(
	provideContext, logger.NewZapLogger, provideCategoryService, handler.NewCategoryHTTP,
)

var transportProxySet = wire.NewSet(
	httpCategorySet,
	provideHandlers, config.NewKernel, transport.NewHTTPServer, transport.NewProxy,
)

func SetContext(rootCtx context.Context) {
	ctx = rootCtx
}

func provideContext() context.Context {
	return ctx
}

func provideCategoryService(ctx2 context.Context, logger2 log.Logger) (service.Category, func(), error) {
	dependency.SetContext(ctx2)

	useCase, cleanup, err := dependency.InjectCategoryUseCase()
	svc := middleware.WrapCategoryMiddleware(useCase, logger2)

	return svc, cleanup, err
}

func provideHandlers(category *handler.CategoryHTTP) []transport.Handler {
	return []transport.Handler{category}
}

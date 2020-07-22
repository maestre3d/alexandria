//+build wireinject

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

var ctx = context.Background()

var httpCategorySet = wire.NewSet(
	provideContext,
	logger.NewZapLogger,
	provideCategoryService,
	handler.NewCategoryHTTP,
)

var transportProxySet = wire.NewSet(
	httpCategorySet,
	provideHandlers,
	config.NewKernel,
	transport.NewHTTPServer,
	transport.NewProxy,
)

func SetContext(rootCtx context.Context) {
	ctx = rootCtx
}

func provideContext() context.Context {
	return ctx
}

func provideCategoryService(ctx context.Context, logger log.Logger) (service.Category, func(), error) {
	dependency.SetContext(ctx)

	useCase, cleanup, err := dependency.InjectCategoryUseCase()
	svc := middleware.WrapCategoryMiddleware(useCase, logger)

	return svc, cleanup, err
}

func provideHandlers(category *handler.CategoryHTTP) []transport.Handler {
	return []transport.Handler{category}
}

func InjectTransportProxy() (*transport.Proxy, func(), error) {
	wire.Build(transportProxySet)
	return &transport.Proxy{}, nil, nil
}

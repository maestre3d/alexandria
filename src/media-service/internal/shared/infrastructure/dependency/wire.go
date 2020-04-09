// +build wireinject

package dependency

import (
	"context"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/application"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	book_infrastructure "github.com/maestre3d/alexandria/src/media-service/internal/media/infrastructure"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/infrastructure"
	"github.com/maestre3d/alexandria/src/media-service/pkg/service/delivery"
	"github.com/maestre3d/alexandria/src/media-service/pkg/service/delivery/handler"
)

var loggerSet = wire.NewSet(
	infrastructure.NewLogger,
	wire.Bind(new(util.ILogger), new(*infrastructure.Logger)),
)
var postgresPoolSet = wire.NewSet(
	ProvideContext,
	loggerSet,
	infrastructure.NewPostgresPool,
)
var bookRepository = wire.NewSet(
	postgresPoolSet,
	book_infrastructure.NewBookRDBMSRepository,
	wire.Bind(new(domain.IBookRepository), new(*book_infrastructure.BookRDBMSRepository)),
)
var bookUseCaseSet = wire.NewSet(
	bookRepository,
	application.NewBookUseCase,
)
var bookHandlerSet = wire.NewSet(
	bookUseCaseSet,
	handler.NewBookHandler,
)
var proxyHandlersSet = wire.NewSet(
	bookHandlerSet,
	ProvideProxyHandlers,
)

func ProvideContext() context.Context {
	return context.Background()
}

func ProvideBookLocalRepository(logger util.ILogger) *book_infrastructure.BookLocalRepository {
	return book_infrastructure.NewBookLocalRepository(make([]*domain.BookEntity, 0), logger)
}

func ProvideProxyHandlers(book *handler.BookHandler) *delivery.ProxyHandlers {
	// Map handlers to proxy
	return &delivery.ProxyHandlers{
		book,
	}
}

func InitHTTPServiceProxy() (*delivery.HTTPServiceProxy, func(), error) {
	wire.Build(wire.NewSet(
		proxyHandlersSet,
		delivery.NewHTTPServer,
		delivery.NewHTTPServiceProxy,
	))

	return &delivery.HTTPServiceProxy{}, nil, nil
}

// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package dependency

import (
	"context"
	"github.com/google/wire"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/application"
	"github.com/maestre3d/alexandria/src/book-service/internal/book/domain"
	infrastructure2 "github.com/maestre3d/alexandria/src/book-service/internal/book/infrastructure"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/infrastructure"
	"github.com/maestre3d/alexandria/src/book-service/pkg/service/delivery"
	"github.com/maestre3d/alexandria/src/book-service/pkg/service/delivery/handler"
)

// Injectors from wire.go:

func InitHTTPServiceProxy() (*delivery.HTTPServiceProxy, func(), error) {
	logger, cleanup, err := infrastructure.NewLogger()
	if err != nil {
		return nil, nil, err
	}
	server := delivery.NewHTTPServer(logger)
	context := ProvideContext()
	db, cleanup2, err := infrastructure.NewPostgresPool(context, logger)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	bookRDBMSRepository := infrastructure2.NewBookRDBMSRepository(db, logger, context)
	bookUseCase := application.NewBookUseCase(logger, bookRDBMSRepository)
	bookHandler := handler.NewBookHandler(logger, bookUseCase)
	proxyHandlers := ProvideProxyHandlers(bookHandler)
	httpServiceProxy := delivery.NewHTTPServiceProxy(logger, server, proxyHandlers)
	return httpServiceProxy, func() {
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var loggerSet = wire.NewSet(infrastructure.NewLogger, wire.Bind(new(util.ILogger), new(*infrastructure.Logger)))

var postgresPoolSet = wire.NewSet(
	ProvideContext,
	loggerSet, infrastructure.NewPostgresPool,
)

var bookRepository = wire.NewSet(
	postgresPoolSet, infrastructure2.NewBookRDBMSRepository, wire.Bind(new(domain.IBookRepository), new(*infrastructure2.BookRDBMSRepository)),
)

var bookUseCaseSet = wire.NewSet(
	bookRepository, application.NewBookUseCase,
)

var bookHandlerSet = wire.NewSet(
	bookUseCaseSet, handler.NewBookHandler,
)

var proxyHandlersSet = wire.NewSet(
	bookHandlerSet,
	ProvideProxyHandlers,
)

func ProvideContext() context.Context {
	return context.Background()
}

func ProvideBookLocalRepository(logger util.ILogger) *infrastructure2.BookLocalRepository {
	return infrastructure2.NewBookLocalRepository(make([]*domain.BookEntity, 0), logger)
}

func ProvideProxyHandlers(book *handler.BookHandler) *delivery.ProxyHandlers {

	return &delivery.ProxyHandlers{
		book,
	}
}

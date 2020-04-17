package main

import (
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/dependency"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"github.com/maestre3d/alexandria/author-service/pkg/shared"
	"github.com/maestre3d/alexandria/author-service/pkg/transport"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/handler"
	"os"
)

func main() {
	authorUsecase, cleanup, err := dependency.ProvideAuthorUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()
	logger := log.NewLogfmtLogger(os.Stderr)

	authorService := author.NewAuthorService(authorUsecase, logger)
	authorHandler := handler.NewAuthorHandler(authorService, logger)

	handlers := &transport.ProxyHandlers{authorHandler}

	server := shared.NewHTTPServer(logger)

	proxyHTTP, cleanup2 := transport.NewHTTPTransportProxy(logger, server, handlers)
	defer cleanup2()

	err = proxyHTTP.Server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

package main

import (
	"context"
	"github.com/maestre3d/alexandria/src/book-service/internal/presentation/delivery"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/infrastructure"
)

func main() {
	// TODO: Use Google's wire DI to correctly inject dependencies
	logger := infrastructure.NewLogger()
	defer logger.Close()
	server := delivery.NewHTTPService(logger)

	err := server.ListenAndServe()
	defer func() {
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	if err != nil {
		panic(err)
	}
}

package main

import (
	"context"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/infrastructure/dependency"
)

func main() {
	httpService, clean, err := dependency.InitHTTPServiceProxy()
	if err != nil {
		panic(err)
	}
	defer clean()

	err = httpService.Server.ListenAndServe()
	defer func() {
		err := httpService.Server.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	if err != nil {
		panic(err)
	}
}

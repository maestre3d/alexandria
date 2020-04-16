package main

import (
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/dependency"
	"github.com/maestre3d/alexandria/author-service/pkg/author"
	"net/http"
	"os"
)

func main() {
	authorSvc, cleanup, err := dependency.InjectAuthorService()
	if err != nil {
		panic(err)
	}
	defer cleanup()
	logger := log.NewLogfmtLogger(os.Stderr)

	svc := author.NewAuthorService(authorSvc, logger)
	r := author.NewTransportHTTP(svc, logger)
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}

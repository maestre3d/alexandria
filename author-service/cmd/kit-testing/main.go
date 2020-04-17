package main

import (
	"github.com/maestre3d/alexandria/author-service/pkg/shared/di"
)

func main() {
	proxyHTTP, cleanup, err := di.InjectHTTPProxy()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	err = proxyHTTP.Server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

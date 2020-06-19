package main

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/media-service/internal/dependency"
	"log"
)

func main() {
	ctx := context.Background()
	dependency.Ctx = ctx
	mediaUse, cleanup, err := dependency.InjectMediaUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	medias, next, err := mediaUse.List(ctx, "", "1", core.FilterParams{
		"filter_by": "timestamp",
	})
	if err != nil {
		panic(err)
	}

	for _, media := range medias {
		log.Printf("%+v", media)
	}
	log.Print(next)
}

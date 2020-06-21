package main

import (
	"context"
	"github.com/maestre3d/alexandria/media-service/internal/dependency"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
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

	media, err := mediaUse.Create(ctx, &domain.MediaAggregate{
		Title:        "1984: Maestre",
		DisplayName:  "1984",
		Description:  "By George Orwell, 1984 is a novel about ... ",
		LanguageCode: "en",
		PublisherID:  "123",
		AuthorID:     "987",
		PublishDate:  "1954-01-31",
		MediaType:    "book",
	})
	if err != nil {
		panic(err)
	}

	log.Printf("%+v", media)
}

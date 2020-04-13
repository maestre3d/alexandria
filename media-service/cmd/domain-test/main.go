package main

import (
	"errors"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/application"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/infrastructure"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/infrastructure/logging"
	"log"
)

func main() {
	logger, cleanup, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	repository := infrastructure.NewMediaLocalRepository(make([]*domain.MediaAggregate, 0), logger)
	usecase := application.NewMediaUseCase(logger, repository)

	params := &application.MediaParams{
		MediaID:     "13",
		Title:       "Thug Life",
		DisplayName: "Green Mille by far",
		Description: "Stephen King is the master of horror stories, fuck you",
		UserID:      "60f90323-fc78-45e4-a0f5-71b63dd87d1a",
		AuthorID:    "a38d10fa-f369-4e8c-8c9d-f7f9f22bdc71",
		PublishDate: "2006-01-31",
		MediaType:   "media_book",
	}

	err = usecase.Create(params)
	if err != nil {
		if errors.Is(err, global.EntityExists) {
			log.Print("exists catch")
		}
		log.Print(err)
		return
	}

	log.Print("media created")

	medias, err := usecase.GetAll(nil, util.FilterParams{})
	if err != nil {
		log.Print(err)
		return
	}

	for _, media := range medias {
		log.Printf("%v", media)
	}
}

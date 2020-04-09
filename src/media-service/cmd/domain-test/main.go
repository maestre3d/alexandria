package main

import (
	"context"
	"errors"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/src/media-service/internal/media/infrastructure"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/global"
	infrastructure2 "github.com/maestre3d/alexandria/src/media-service/internal/shared/infrastructure"
	"go.uber.org/multierr"
	"log"
	"strings"
)

func main() {
	logger, cleanup, err := infrastructure2.NewLogger()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	db, cleanup, err := infrastructure2.NewPostgresPool(context.Background(), logger)
	defer cleanup()

	repository := infrastructure.NewMediaRDBMSRepository(db, logger, context.Background())

	params := &domain.MediaEntityParams{
		Title:       "Green Mille",
		DisplayName: "Green Mille by far",
		Description: "Stephen King is the master of horror stories, fuck you",
		UserID:      "60f90323-fc78-45e4-a0f5-71b63dd87d1a",
		AuthorID:    "a38d10fa-f369-4e8c-8c9d-f7f9f22bdc71",
		PublishDate: "2006-01-31",
		MediaType:   "media_book",
	}

	media, err := domain.NewMediaEntity(params)
	errs := multierr.Errors(err)
	if len(errs) > 0 {
		for _, err = range errs {
			errDesc := strings.Split(err.Error(), ":")
			if len(errDesc) > 1 {
				if errors.Is(err, global.InvalidFieldFormat) {
					//log.Print(errors.Unwrap(err))
					log.Print(errDesc[1])
					return
				}

				log.Print(errDesc[1])
				return
			}

			log.Print(err)
			return
		}
	}

	mediaAg := media.ToMediaAggregate()
	err = repository.Save(mediaAg)
	if err != nil {
		log.Print(errors.Unwrap(err))
	}

	log.Print("saved media record")

	mediasFromRep, err := repository.Fetch(nil)
	if err != nil {
		if errors.Is(err, global.EntitiesNotFound) {
			logger.Print(err.Error(), "main")
			return
		}

		log.Print(errors.Unwrap(err))
		return
	}

	for _, mediaRep := range mediasFromRep {
		log.Printf("%v", mediaRep)
	}

}

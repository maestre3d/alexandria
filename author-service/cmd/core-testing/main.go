package main

import (
	"errors"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/dependency"
	"log"
)

func main() {
	log.Print("Hello from author service")
	authorService, cleanup, err := dependency.InjectAuthorUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()
	/*
		authorCreated, err := authorService.Create("Joel", "Coen", "Joel Coen", "1975-06-30")
		if err != nil {
			if errors.Is(err, exception.RequiredField) || errors.Is(err, exception.InvalidFieldRange) || errors.Is(err, exception.InvalidFieldFormat) {
				// 400 HTTP Error
				errDesc := strings.Split(err.Error(), ":")
				if len(errDesc) > 1 {
					log.Print(errDesc[1])
					return
				}
			} else if errors.Is(err, exception.EntityExists) {
				// 409 HTTP Error
			}

			// 500 HTTP Error
			log.Print(err)
			return
		}

		log.Printf("%v", authorCreated)
	*/

	authorGet, err := authorService.Get("2590a2c6-7692-4e09-92c8-427f8e3824cf")
	if err != nil {
		if errors.Is(err, exception.InvalidID) {
			// 400 HTTP Error
			log.Print(err)
			return
		}

		log.Print(err)
		return
	}

	if authorGet == nil {
		// 404 HTTP Error
		log.Print("author not found")
	}

	log.Printf("%v", authorGet)

	filterParams := util.FilterParams{
		"query": "Ethan",
	}

	authors, _ , err := authorService.List("", "0", filterParams)
	if err != nil {
		log.Print(err)
		return
	}

	if len(authors) == 0 {
		// 404 HTTP Error
		log.Print("authors not found")
	}

	log.Printf("%v", authors[0])
}

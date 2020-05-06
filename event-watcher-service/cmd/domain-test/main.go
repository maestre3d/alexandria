package main

import (
	"github.com/maestre3d/alexandria/event-watcher-service/internal/shared/infrastructure/dependency"
	"log"
)

func main() {
	eventUseCase, cleanup, err := dependency.InjectEventUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// Create

	/*
		contentExample := struct {
			Message string
		}{
			Message: "hello there",
		}

		exampleJSON, err := json.Marshal(contentExample)
		if err != nil {
			panic(err)
		}

		eventCreated, err := eventUseCase.Create("foo", "", "domain", string(exampleJSON), "low", "kafka")
		if err != nil {
			panic(err)
		}

		// Get
		event, err := eventUseCase.Get(fmt.Sprintf(`%s#%d`, eventCreated.ID, eventCreated.DispatchTime))
		if err != nil {
			panic(err)
		}
		log.Printf("%v", event)*/

	/*
		// Update
		contentExample.Message = "Hello there updated"

		exampleJSON, err = json.Marshal(contentExample)
		if err != nil {
			panic(err)
		}

		eventUpdated, err := eventUseCase.Update(fmt.Sprintf(`%s#%d`, event.ID, event.DispatchTime), "bar", "",
			"interaction", string(exampleJSON), "low", "rabbitmq")
		if err != nil {
			panic(err)
		}
		log.Printf("%v", eventUpdated)

		// Delete
		if err := eventUseCase.Delete(fmt.Sprintf(`%s#%d`, event.ID, event.DispatchTime)); err != nil {
			panic(err)
		}
		log.Print("event deleted ")*/

	// List
	events, next, err := eventUseCase.List("", "1", nil)
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		log.Printf("%v", event)
	}
	log.Print(next)
}
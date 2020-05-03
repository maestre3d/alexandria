package main

import (
	"log"

	"github.com/maestre3d/alexandria/event-watcher-service/internal/shared/infrastructure/dependency"
)

func main() {
	watcherUseCase, cleanup, err := dependency.InjectWatcherUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	watcherCreated, err := watcherUseCase.Create("foo", "", "domain", "", "low", "kafka")
	if err != nil {
		panic(err)
	}

	log.Printf("%v", watcherCreated)
}

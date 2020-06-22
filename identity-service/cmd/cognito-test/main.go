package main

import (
	"context"
	"github.com/maestre3d/alexandria/identity-service/internal/dependency"
	"log"
)

func main() {
	ctx := context.Background()
	dependency.Ctx = ctx
	userCase, err := dependency.InjectUserUseCase()
	if err != nil {
		panic(err)
	}

	user, err := userCase.Get(ctx, "69817804-4af4-4de1-83af-4a5f660d0018")
	if err != nil {
		panic(err)
	}

	log.Print(user)
}

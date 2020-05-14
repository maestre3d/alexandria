package main

import (
	"github.com/maestre3d/alexandria/identity-service/internal/infrastructure/dependency"
	"log"
)

func main() {
	userUseCase, cleanup, err := dependency.InjectUserUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	user, err := userUseCase.Create("aruizmx", "alonso", "ruiz", "aruiz1@gmail.com", "male", "")
	if err != nil {
		panic(err)
	}

	log.Print(user)
}

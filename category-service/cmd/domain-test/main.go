package main

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/category-service/internal/dependency"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"log"
	"sync"
)

func main() {
	ctx := context.Background()

	dependency.SetContext(ctx)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	ctx2, _ := context.WithCancel(ctx)
	go TestCategoryRoot(ctx2, wg)
	ctx3, _ := context.WithCancel(ctx)
	go TestCategory(ctx3, wg)

	wg.Wait()
}

func TestCategory(ctx context.Context, wg *sync.WaitGroup) {
	categoryI, cleanup, err := dependency.InjectCategoryUseCase()
	if err != nil {
		log.Panic(err)
	}
	defer cleanup()

	category := new(domain.Category)
	catChan := make(chan *domain.Category)
	go func() {
		ctx1, _ := context.WithCancel(ctx)
		category, err := categoryI.Create(ctx1, "comedy")
		if err != nil {
			log.Print(err)
			return
		}
		catChan <- category
	}()
	select {
	case category = <-catChan:
		log.Printf("%+v", category)
		break
	}

	ctx2, _ := context.WithCancel(ctx)
	err = categoryI.HardDelete(ctx2, category.ExternalID)
	if err != nil {
		log.Print(err)
		return
	}

	ctx3, _ := context.WithCancel(ctx)
	categories, token, err := categoryI.List(ctx3, "", "10", core.FilterParams{"query": "sci-fi"})
	if err != nil {
		log.Print(err)
		return
	}

	for _, cat := range categories {
		log.Printf("%+v", cat)
	}

	log.Printf("next_token: %s", token)
	wg.Done()
}

func TestCategoryRoot(ctx context.Context, wg *sync.WaitGroup) {
	categoryI, cleanup, err := dependency.InjectCategoryRootUseCase()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// Mock UUID
	rootID := uuid.New().String()
	ctx1, _ := context.WithCancel(ctx)
	list, err := categoryI.CreateList(ctx1, "766VUh8YnfHBJtv-", rootID)
	if err != nil {
		log.Print(err)
		return
	}

	log.Printf("%+v", list)

	ctx2, _ := context.WithCancel(ctx)
	categoryID := "SSeaU0gfoVTJfCqk"
	err = categoryI.Add(ctx2, categoryID, rootID)
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("added %s category to %s", categoryID, rootID)

	ctx3, _ := context.WithCancel(ctx)
	err = categoryI.DeleteItem(ctx3, rootID, categoryID)
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("removed category %s from %s", categoryID, rootID)

	ctx4, _ := context.WithCancel(ctx)
	listRoot, err := categoryI.GetByRoot(ctx4, rootID)
	if err != nil {
		log.Print(err)
		return
	}

	for id, name := range listRoot.CategoryList {
		log.Printf("category id: %s, category name: %s", id, name)
	}

	ctx5, _ := context.WithCancel(ctx)
	err = categoryI.DeleteList(ctx5, rootID)
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("list from %s removed", rootID)

	wg.Done()
}

package main

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/internal/infrastructure"
	"github.com/maestre3d/alexandria/category-service/internal/interactor"
)

func main() {
	ctx := context.Background()
	log := logger.NewZapLogger()
	cfg, err := config.NewKernel(ctx)
	if err != nil {
		panic(err)
	}
	redis, cleanR, err := persistence.NewRedisPool(cfg)
	if err != nil {
		panic(err)
	}
	defer cleanR()
	cassandra := infrastructure.NewCassandraPool(cfg)

	repo := infrastructure.NewCategoryRepositoryCassandra(cassandra, redis)
	eventBus := infrastructure.NewCategoryEventKafka(cfg)
	categoryI := interactor.NewCategoryUseCase(log, repo, eventBus)

	category := new(domain.Category)
	catChan := make(chan *domain.Category)
	go func() {
		ctx1, _ := context.WithCancel(ctx)
		category, err := categoryI.Create(ctx1, "sci-fi")
		if err != nil {
			panic(err)
		}
		catChan <- category
	}()
	select {
	case category = <-catChan:
		_ = log.Log("msg", fmt.Sprintf("%+v", category))
		break
	}

	ctx2, _ := context.WithCancel(ctx)
	err = categoryI.HardDelete(ctx2, category.ExternalID)
	if err != nil {
		_ = log.Log("err", err)
		panic(err)
	}

	ctx3, _ := context.WithCancel(ctx)
	categories, token, err := categoryI.List(ctx3, "", "10", core.FilterParams{"query": "terror"})
	if err != nil {
		panic(err)
	}

	for _, cat := range categories {
		_ = log.Log("output", fmt.Sprintf("%+v", cat))
	}

	_ = log.Log("next_token", token)
}

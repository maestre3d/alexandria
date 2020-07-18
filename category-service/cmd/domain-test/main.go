package main

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/logger"
	"github.com/alexandria-oss/core/persistence"
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

	/*
		createdCat, err := categoryI.Create(ctx, "terror")
		if err != nil {
			panic(err)
		}


	*/

	createdCat, token, err := categoryI.List(ctx, "", "1", core.FilterParams{"query": "terror"})
	if err != nil {
		panic(err)
	}

	for _, cat := range createdCat {
		_ = log.Log("output", fmt.Sprintf("%+v", cat))
	}

	_ = log.Log("next_token", token)
}

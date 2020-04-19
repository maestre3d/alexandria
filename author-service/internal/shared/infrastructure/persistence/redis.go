package persistence

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/config"
)

func NewRedisPool(logger log.Logger, cfg *config.KernelConfig) (*redis.Client, func(), error) {
	defer func(begin time.Time) {
		logger.Log(
			"method", "core.kernel.infrastructure.persistence",
			"msg", "in-memory database started",
			"took", time.Since(begin),
		)
	}(time.Now())

	var db int
	var err error

	db, err = strconv.Atoi(cfg.MemConfig.Database)
	if err != nil {
		db = 0
	}

	client := redis.NewClient(&redis.Options{
		Network:            cfg.MemConfig.Network,
		Addr:               cfg.MemConfig.Host + fmt.Sprintf(":%d", cfg.MemConfig.Port),
		Dialer:             nil,
		OnConnect:          nil,
		Password:           cfg.MemConfig.Password,
		DB:                 db,
		MaxRetries:         10,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        30 * time.Second,
		ReadTimeout:        15 * time.Second,
		WriteTimeout:       15 * time.Second,
		PoolSize:           100,
		MinIdleConns:       32,
		MaxConnAge:         0,
		PoolTimeout:        24 * time.Second,
		IdleTimeout:        30 * time.Second,
		IdleCheckFrequency: 0,
		TLSConfig:          nil,
		Limiter:            nil,
	})

	cleanup := func() {
		if client != nil {
			client.Close()
		}
	}

	err = client.Ping().Err()
	if err != nil {
		return nil, cleanup, nil
	}

	return client, cleanup, nil
}

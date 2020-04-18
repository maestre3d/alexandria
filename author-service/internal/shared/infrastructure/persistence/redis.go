package persistence

import (
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/config"
	"time"
)

func NewRedisPool(logger log.Logger, cfg *config.KernelConfig) (*redis.Client, func(), error) {
	defer func(begin time.Time) {
		logger.Log(
			"method", "core.kernel.infrastructure.persistence",
			"msg", "in-memory database started",
			"took", time.Since(begin),
		)
	}(time.Now())

	client := redis.NewClient(&redis.Options{
		Network:            "",
		Addr:               cfg.MainMemHost,
		Dialer:             nil,
		OnConnect:          nil,
		Password:           cfg.MainMemPassword,
		DB:                 0,
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

	err := client.Ping().Err()
	if err != nil {
		return nil, cleanup, nil
	}

	return client, cleanup, nil
}

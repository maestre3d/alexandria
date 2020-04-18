package persistence

import (
	"github.com/go-redis/redis/v7"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
)

func NewRedisPool(logger util.ILogger) (*redis.Client, func(), error) {
	client := redis.NewClient(&redis.Options{
		Network:            "",
		Addr:               "localhost:6379",
		Dialer:             nil,
		OnConnect:          nil,
		Password:           "",
		DB:                 0,
		MaxRetries:         0,
		MinRetryBackoff:    0,
		MaxRetryBackoff:    0,
		DialTimeout:        0,
		ReadTimeout:        0,
		WriteTimeout:       0,
		PoolSize:           0,
		MinIdleConns:       0,
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
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
	logger.Print("in-memory database started", "kernel.infrastructure.persistence")

	return client, cleanup, nil
}

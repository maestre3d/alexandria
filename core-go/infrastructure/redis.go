package infrastructure

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/maestre3d/alexandria/core-go/config"
)

// NewRedisPool Obtain a Redis connection pool
func NewRedisPool(cfg *config.KernelConfiguration) (*redis.Client, func(), error) {
	db, err := strconv.Atoi(cfg.InMemoryConfig.Database)
	if err != nil {
		db = 0
	}

	client := redis.NewClient(&redis.Options{
		Network:            cfg.InMemoryConfig.Network,
		Addr:               cfg.InMemoryConfig.Host + fmt.Sprintf(":%d", cfg.InMemoryConfig.Port),
		Dialer:             nil,
		OnConnect:          nil,
		Password:           cfg.InMemoryConfig.Password,
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

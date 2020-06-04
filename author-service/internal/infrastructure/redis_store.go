package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

// Store a record inside redis database using the write-through pattern, ignore errors
// (recommended for optional cache)
func Store(ctx context.Context, c *redis.Client, id, key string, entity interface{}) {
	memConn := c.Conn()
	defer func() {
		_ = memConn.Close()
	}()

	entityJSON, err := json.Marshal(entity)
	if err == nil {
		err = memConn.Set(fmt.Sprintf("%s:%s", key, id), entityJSON, (time.Hour * 1)).Err()
	}
}

// StoreSafe a record inside redis database using the write-through pattern
func StoreSafe(ctx context.Context, c *redis.Client, id, key string, entity interface{}) error {
	memConn := c.Conn()
	defer func() {
		_ = memConn.Close()
	}()

	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	return memConn.Set(fmt.Sprintf("%s:%s", key, id), entityJSON, (time.Hour * 1)).Err()
}

// Get returns an entity from redis store using the write-through pattern, ignores errors
// (recommended for optional cache)
func Get(ctx context.Context, c *redis.Client, id, key string) interface{} {
	memCon := c.Conn()
	defer func() {
		err := memCon.Close()
		if err != nil {
			return
		}
	}()

	entityMem, err := memCon.Get(fmt.Sprintf("%s:%s", key, id)).Result()
	if err == nil {
		entity := &struct{}{}
		err = json.Unmarshal([]byte(entityMem), entity)
		if err == nil {
			return entity
		}

		return nil
	}

	return nil
}

// GetSafe returns an entity from redis store using the write-through pattern
func GetSafe(ctx context.Context, c *redis.Client, id, key string) (interface{}, error) {
	memCon := c.Conn()
	defer func() {
		err := memCon.Close()
		if err != nil {
			return
		}
	}()

	entityMem, err := memCon.Get(fmt.Sprintf("%s:%s", key, id)).Result()
	if err != nil {
		return nil, err
	}

	entity := &struct{}{}
	err = json.Unmarshal([]byte(entityMem), entity)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// Remove an entity from a store
func Remove(ctx context.Context, c *redis.Client, id, key string) {
	memCon := c.Conn()
	defer func() {
		err := memCon.Close()
		if err != nil {
			return
		}
	}()

	err := memCon.Get(fmt.Sprintf("%s:%s", key, id)).Err()
	if err == nil {
		err = memCon.Del(key + ":" + id).Err()
	}
}

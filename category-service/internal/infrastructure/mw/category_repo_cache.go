package mw

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/config"
	"github.com/go-redis/redis/v7"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"strings"
	"sync"
	"time"
)

type CategoryRepositoryCache struct {
	Pool *redis.Client
	Cfg  *config.Kernel
	Next domain.CategoryRepository
	Mu   *sync.RWMutex
}

func (c CategoryRepositoryCache) Save(ctx context.Context, category domain.Category) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	return c.Next.Save(ctx, category)
}

func (c CategoryRepositoryCache) FetchByID(ctx context.Context, id string, activeOnly bool) (category *domain.Category, err error) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()

	// Before call, fetch from cache first
	var conn *redis.Conn
	var memKey string
	if c.Pool != nil {
		memKey = fmt.Sprintf("%s:%s", strings.ToLower(c.Cfg.Service), id)

		conn = c.Pool.Conn()
		defer func() {
			_ = conn.Close()
		}()

		// If error is found, then proceed to read from main database
		// Entity structure -> service_name:json_entity
		if categoryJSON, err := conn.Get(memKey).Result(); err == nil {
			category = new(domain.Category)
			err = json.Unmarshal([]byte(categoryJSON), category)
			if err == nil {
				return category, nil
			}
		}
	}

	category, err = c.Next.FetchByID(ctx, id, activeOnly)
	// If entity found, then store it to cache
	if category != nil {
		if c.Pool != nil {
			if errR := conn.Del(memKey).Err(); errR == nil {
				if categoryJSON, errJ := json.Marshal(category); errJ == nil {
					_ = conn.Set("", categoryJSON, 60*time.Minute)
				}
			}
		}
	}

	return
}

func (c CategoryRepositoryCache) Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*domain.Category, error) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return c.Next.Fetch(ctx, params, filter)
}

func (c CategoryRepositoryCache) Replace(ctx context.Context, category domain.Category) (err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	err = c.Next.Replace(ctx, category)
	if err == nil {
		// If processing has no errors, then invalidate cache
		if c.Pool != nil {
			conn := c.Pool.Conn()
			defer func() {
				_ = conn.Close()
			}()

			_ = conn.Del(fmt.Sprintf("%s:%s", strings.ToLower(c.Cfg.Service), category.ExternalID)).Err()
		}
	}
	return
}

func (c CategoryRepositoryCache) Remove(ctx context.Context, id string) (err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	err = c.Next.Remove(ctx, id)
	if err == nil {
		// If processing has no errors, then invalidate cache
		if c.Pool != nil {
			conn := c.Pool.Conn()
			defer func() {
				_ = conn.Close()
			}()

			_ = conn.Del(fmt.Sprintf("%s:%s", strings.ToLower(c.Cfg.Service), id)).Err()
		}
	}

	return
}

func (c CategoryRepositoryCache) Restore(ctx context.Context, id string) (err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	err = c.Next.Restore(ctx, id)
	if err == nil {
		// If processing has no errors, then invalidate cache
		if c.Pool != nil {
			conn := c.Pool.Conn()
			defer func() {
				_ = conn.Close()
			}()

			_ = conn.Del(fmt.Sprintf("%s:%s", strings.ToLower(c.Cfg.Service), id)).Err()
		}
	}

	return
}

func (c CategoryRepositoryCache) HardRemove(ctx context.Context, id string) (err error) {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	err = c.Next.HardRemove(ctx, id)
	if err == nil {
		// If processing has no errors, then invalidate cache
		if c.Pool != nil {
			conn := c.Pool.Conn()
			defer func() {
				_ = conn.Close()
			}()

			_ = conn.Del(fmt.Sprintf("%s:%s", strings.ToLower(c.Cfg.Service), id)).Err()
		}
	}

	return
}

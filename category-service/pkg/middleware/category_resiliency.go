package middleware

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/pkg/service"
	"go.uber.org/ratelimit"
)

type CategoryResiliency struct {
	Logger      log.Logger
	RateLimiter ratelimit.Limiter
	Next        service.Category
}

func (r CategoryResiliency) Create(ctx context.Context, name string) (*domain.Category, error) {
	r.RateLimiter.Take()
	return r.Next.Create(ctx, name)
}

func (r CategoryResiliency) Get(ctx context.Context, id string) (*domain.Category, error) {
	r.RateLimiter.Take()
	return r.Next.Get(ctx, id)
}

func (r CategoryResiliency) List(ctx context.Context, token, limit string,
	filter core.FilterParams) ([]*domain.Category, string, error) {
	r.RateLimiter.Take()
	return r.Next.List(ctx, token, limit, filter)
}

func (r CategoryResiliency) Update(ctx context.Context, id string, name string) (*domain.Category, error) {
	r.RateLimiter.Take()
	return r.Next.Update(ctx, id, name)
}

func (r CategoryResiliency) Delete(ctx context.Context, id string) error {
	r.RateLimiter.Take()
	return r.Next.Delete(ctx, id)
}

func (r CategoryResiliency) Restore(ctx context.Context, id string) error {
	r.RateLimiter.Take()
	return r.Next.Restore(ctx, id)
}

func (r CategoryResiliency) HardDelete(ctx context.Context, id string) error {
	r.RateLimiter.Take()
	return r.Next.HardDelete(ctx, id)
}

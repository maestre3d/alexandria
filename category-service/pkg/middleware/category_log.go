package middleware

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/pkg/service"
	"time"
)

type CategoryLog struct {
	Logger log.Logger
	Next   service.Category
}

func (l CategoryLog) Create(ctx context.Context, name string) (category *domain.Category, err error) {
	defer func(begin time.Time) {
		_ = level.Info(l.Logger).Log(
			"endpoint", "category.create",
			"input", fmt.Sprintf("name: %s", name),
			"output", fmt.Sprintf("category: %+v", category),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	category, err = l.Next.Create(ctx, name)
	return
}

func (l CategoryLog) Get(ctx context.Context, id string) (category *domain.Category, err error) {
	defer func(begin time.Time) {
		_ = level.Info(l.Logger).Log(
			"endpoint", "category.get",
			"input", fmt.Sprintf("id: %s", id),
			"output", fmt.Sprintf("category: %+v", category),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	category, err = l.Next.Get(ctx, id)
	return
}

func (l CategoryLog) List(ctx context.Context, token, limit string,
	filter core.FilterParams) (categories []*domain.Category, nextToken string, err error) {
	defer func(begin time.Time) {
		_ = level.Info(l.Logger).Log(
			"endpoint", "category.list",
			"input", fmt.Sprintf("token: %s, limit: %s, filter: %v", token, limit, filter),
			"output", fmt.Sprintf("categories: %+v, next_token: %s", categories, nextToken),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	categories, nextToken, err = l.Next.List(ctx, token, limit, filter)
	return
}

func (l CategoryLog) Update(ctx context.Context, id string, name string) (category *domain.Category, err error) {
	defer func(begin time.Time) {
		_ = level.Info(l.Logger).Log(
			"endpoint", "category.update",
			"input", fmt.Sprintf("id: %s, name: %s", id, name),
			"output", fmt.Sprintf("category: %+v", category),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	category, err = l.Next.Update(ctx, id, name)
	return
}

func (l CategoryLog) Delete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		_ = level.Info(l.Logger).Log(
			"endpoint", "category.delete",
			"input", fmt.Sprintf("id: %s", id),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = l.Next.Delete(ctx, id)
	return
}

func (l CategoryLog) Restore(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		_ = level.Info(l.Logger).Log(
			"endpoint", "category.restore",
			"input", fmt.Sprintf("id: %s", id),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = l.Next.Restore(ctx, id)
	return
}

func (l CategoryLog) HardDelete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		_ = level.Info(l.Logger).Log(
			"endpoint", "category.hard_delete",
			"input", fmt.Sprintf("id: %s", id),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	err = l.Next.HardDelete(ctx, id)
	return
}

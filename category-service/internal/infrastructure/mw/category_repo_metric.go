package mw

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"time"
)

type CategoryRepositoryMetric struct {
	Next domain.CategoryRepository
}

var (
	latencyMs    = stats.Float64("category/cassandra/latency", "The latency in milliseconds per Apache Cassandra read", "ms")
	keyMethod, _ = tag.NewKey("method")
	keyStatus, _ = tag.NewKey("status")
	keyError, _  = tag.NewKey("error")
)

func (c CategoryRepositoryMetric) Save(ctx context.Context, category domain.Category) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(begin).Nanoseconds())/1e6))
	}(time.Now())

	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.save"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM

	err = c.Next.Save(ctx, category)
	return
}

func (c CategoryRepositoryMetric) Fetch(ctx context.Context, params core.PaginationParams,
	filter core.FilterParams) (categories []*domain.Category, err error) {
	defer func(begin time.Time) {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(begin).Nanoseconds())/1e6))
	}(time.Now())

	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.fetch"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return nil, err
	}
	ctx = ctxM

	categories, err = c.Next.Fetch(ctx, params, filter)
	return
}

func (c CategoryRepositoryMetric) FetchByID(ctx context.Context, id string, activeOnly bool) (category *domain.Category, err error) {
	defer func(begin time.Time) {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(begin).Nanoseconds())/1e6))
	}(time.Now())

	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.save"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return nil, err
	}
	ctx = ctxM

	category, err = c.Next.FetchByID(ctx, id, activeOnly)
	return
}

func (c CategoryRepositoryMetric) Replace(ctx context.Context, category domain.Category) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(begin).Nanoseconds())/1e6))
	}(time.Now())

	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.replace"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM

	err = c.Next.Replace(ctx, category)
	return
}

func (c CategoryRepositoryMetric) Remove(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(begin).Nanoseconds())/1e6))
	}(time.Now())

	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.remove"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM

	err = c.Next.Remove(ctx, id)
	return
}

func (c CategoryRepositoryMetric) Restore(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(begin).Nanoseconds())/1e6))
	}(time.Now())

	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.restore"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM

	err = c.Next.Restore(ctx, id)
	return
}

func (c CategoryRepositoryMetric) HardRemove(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(begin).Nanoseconds())/1e6))
	}(time.Now())

	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.hard_remove"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM

	err = c.Next.HardRemove(ctx, id)
	return
}

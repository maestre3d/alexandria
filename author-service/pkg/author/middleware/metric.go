package middleware

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"time"
)

type MetricAuthorMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	Next           usecase.AuthorInteractor
}

func (mw MetricAuthorMiddleware) Create(ctx context.Context, aggregate *domain.AuthorAggregate) (output *domain.Author, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.create", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Create(ctx, aggregate)
	return
}

func (mw MetricAuthorMiddleware) List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) (output []*domain.Author, nextToken string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.list", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, nextToken, err = mw.Next.List(ctx, pageToken, pageSize, filterParams)
	return
}

func (mw MetricAuthorMiddleware) Get(ctx context.Context, id string) (output *domain.Author, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.get", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

func (mw MetricAuthorMiddleware) Update(ctx context.Context, aggregate *domain.AuthorUpdateAggregate) (output *domain.Author, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.update", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Update(ctx, aggregate)
	return
}

func (mw MetricAuthorMiddleware) Delete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.delete", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Delete(ctx, id)
	return
}

func (mw MetricAuthorMiddleware) Restore(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.restore", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Restore(ctx, id)
	return
}

func (mw MetricAuthorMiddleware) HardDelete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.hard_delete", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Delete(ctx, id)
	return
}

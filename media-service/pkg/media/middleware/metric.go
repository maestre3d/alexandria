package middleware

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	"time"
)

type MetricMediaMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	RequestGauge   metrics.Gauge
	Next           usecase.MediaInteractor
}

func (mw MetricMediaMiddleware) Create(ctx context.Context, aggregate *domain.MediaAggregate) (output *domain.Media, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.create", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		mw.RequestGauge.With(lvs...).Add(1)
	}(time.Now())

	output, err = mw.Next.Create(ctx, aggregate)
	return
}

func (mw MetricMediaMiddleware) List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) (output []*domain.Media, nextToken string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.list", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, nextToken, err = mw.Next.List(ctx, pageToken, pageSize, filterParams)
	return
}

func (mw MetricMediaMiddleware) Get(ctx context.Context, id string) (output *domain.Media, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.get", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

func (mw MetricMediaMiddleware) Update(ctx context.Context, aggregate *domain.MediaUpdateAggregate) (output *domain.Media, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.update", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Update(ctx, aggregate)
	return
}

func (mw MetricMediaMiddleware) Delete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.delete", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Delete(ctx, id)
	return
}

func (mw MetricMediaMiddleware) Restore(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.restore", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Restore(ctx, id)
	return
}

func (mw MetricMediaMiddleware) HardDelete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.hard_delete", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.HardDelete(ctx, id)
	return
}

type MetricMediaSAGAMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	RequestGauge   metrics.Gauge
	Next           usecase.MediaSAGAInteractor
}

func (mw MetricMediaSAGAMiddleware) VerifyAuthor(ctx context.Context, rootID string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.saga.verify_author", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.VerifyAuthor(ctx, rootID)
	return
}

func (mw MetricMediaSAGAMiddleware) UpdateStatic(ctx context.Context, rootID string, urlJSON []byte) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.saga.update_static", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.UpdateStatic(ctx, rootID, urlJSON)
	return
}

func (mw MetricMediaSAGAMiddleware) RemoveStatic(ctx context.Context, rootID []byte) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.saga.remove_static", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.RemoveStatic(ctx, rootID)
	return
}

func (mw MetricMediaSAGAMiddleware) Done(ctx context.Context, rootID, operation string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.saga.done", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Done(ctx, rootID, operation)
	return
}

func (mw MetricMediaSAGAMiddleware) Failed(ctx context.Context, rootID, operation, backup string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.saga.failed", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Failed(ctx, rootID, operation, backup)
	return
}

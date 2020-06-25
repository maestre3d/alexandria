package middleware

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	"time"
)

type MetricBlobMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	Next           usecase.BlobInteractor
}

func (mw MetricBlobMiddleware) Store(ctx context.Context, aggregate *domain.BlobAggregate) (output *domain.Blob, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "blob.create", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Store(ctx, aggregate)
	return
}

func (mw MetricBlobMiddleware) Get(ctx context.Context, id, service string) (output *domain.Blob, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "blob.get", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Get(ctx, id, service)
	return
}

func (mw MetricBlobMiddleware) Delete(ctx context.Context, id, service string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "blob.delete", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Delete(ctx, id, service)
	return
}

package middleware

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
	"time"
)

type MetricUserMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	Next           usecase.UserInteractor
}

func (mw MetricUserMiddleware) Get(ctx context.Context, id string) (output *domain.User, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "user.get", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

type MetricUserSAGAMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	Next           usecase.UserSAGAInteractor
}

func (mw MetricUserSAGAMiddleware) Verify(ctx context.Context, usersJSON []byte) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "user.saga.verify", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Verify(ctx, usersJSON)
	return
}

func (mw MetricUserSAGAMiddleware) UpdatePicture(ctx context.Context, id string, urlJSON []byte) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "user.saga.update_picture", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.UpdatePicture(ctx, id, urlJSON)
	return
}

func (mw MetricUserSAGAMiddleware) RemovePicture(ctx context.Context, rootJSON []byte) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "user.saga.remove_picture", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.RemovePicture(ctx, rootJSON)
	return
}

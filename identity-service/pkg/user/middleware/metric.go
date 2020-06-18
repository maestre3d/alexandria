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

package middleware

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/pkg/service"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type CategoryMetric struct {
	ReqCounter      *prometheus.CounterVec
	ReqErrCounter   *prometheus.CounterVec
	ReqHistogram    *prometheus.HistogramVec
	CategoriesTotal prometheus.Gauge
	Next            service.Category
}

func (c CategoryMetric) Create(ctx context.Context, name string) (category *domain.Category, err error) {
	defer func(begin time.Time) {
		lvs := prometheus.Labels{"method": "category.create", "error": fmt.Sprint(err != nil)}
		c.ReqCounter.With(lvs).Inc()
		c.ReqHistogram.With(lvs).Observe(time.Since(begin).Seconds())
		if err != nil {
			c.ReqErrCounter.With(prometheus.Labels{"method": "category.create"}).Inc()
		}

		// Custom metrics
		c.CategoriesTotal.Inc()
	}(time.Now())

	category, err = c.Next.Create(ctx, name)
	return
}

func (c CategoryMetric) Get(ctx context.Context, id string) (category *domain.Category, err error) {
	defer func(begin time.Time) {
		lvs := prometheus.Labels{"method": "category.get", "error": fmt.Sprint(err != nil)}
		c.ReqCounter.With(lvs).Inc()
		c.ReqHistogram.With(lvs).Observe(time.Since(begin).Seconds())
		if err != nil {
			c.ReqErrCounter.With(prometheus.Labels{"method": "category.get"}).Inc()
		}
	}(time.Now())

	category, err = c.Next.Get(ctx, id)
	return
}

func (c CategoryMetric) List(ctx context.Context, token, limit string,
	filter core.FilterParams) (categories []*domain.Category, nextToken string, err error) {
	defer func(begin time.Time) {
		lvs := prometheus.Labels{"method": "category.list", "error": fmt.Sprint(err != nil)}
		c.ReqCounter.With(lvs).Inc()
		c.ReqHistogram.With(lvs).Observe(time.Since(begin).Seconds())
		if err != nil {
			c.ReqErrCounter.With(prometheus.Labels{"method": "category.list"}).Inc()
		}
	}(time.Now())

	categories, nextToken, err = c.Next.List(ctx, token, limit, filter)
	return
}

func (c CategoryMetric) Update(ctx context.Context, id string, name string) (category *domain.Category, err error) {
	defer func(begin time.Time) {
		lvs := prometheus.Labels{"method": "category.update", "error": fmt.Sprint(err != nil)}
		c.ReqCounter.With(lvs).Inc()
		c.ReqHistogram.With(lvs).Observe(time.Since(begin).Seconds())
		if err != nil {
			c.ReqErrCounter.With(prometheus.Labels{"method": "category.update"}).Inc()
		}
	}(time.Now())

	category, err = c.Next.Update(ctx, id, name)
	return
}

func (c CategoryMetric) Delete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := prometheus.Labels{"method": "category.delete", "error": fmt.Sprint(err != nil)}
		c.ReqCounter.With(lvs).Inc()
		c.ReqHistogram.With(lvs).Observe(time.Since(begin).Seconds())
		if err != nil {
			c.ReqErrCounter.With(prometheus.Labels{"method": "category.delete"}).Inc()
		}

		// Custom metrics
		c.CategoriesTotal.Dec()
	}(time.Now())

	err = c.Next.Delete(ctx, id)
	return
}

func (c CategoryMetric) Restore(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := prometheus.Labels{"method": "category.restore", "error": fmt.Sprint(err != nil)}
		c.ReqCounter.With(lvs).Inc()
		c.ReqHistogram.With(lvs).Observe(time.Since(begin).Seconds())
		if err != nil {
			c.ReqErrCounter.With(prometheus.Labels{"method": "category.restore"}).Inc()
		}

		// Custom metrics
		c.CategoriesTotal.Inc()
	}(time.Now())

	err = c.Next.Restore(ctx, id)
	return
}

func (c CategoryMetric) HardDelete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := prometheus.Labels{"method": "category.hard_delete", "error": fmt.Sprint(err != nil)}
		c.ReqCounter.With(lvs).Inc()
		c.ReqHistogram.With(lvs).Observe(time.Since(begin).Seconds())
		if err != nil {
			c.ReqErrCounter.With(prometheus.Labels{"method": "category.hard_delete"}).Inc()
		}

		// Custom metrics
		c.CategoriesTotal.Dec()
	}(time.Now())

	err = c.Next.HardDelete(ctx, id)
	return
}

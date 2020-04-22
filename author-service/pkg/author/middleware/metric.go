package middleware

import (
	"fmt"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"time"
)

type InstrumentingAuthorMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	Next           service.IAuthorService
}

func (mw InstrumentingAuthorMiddleware) Create(firstName, lastName, displayName, birthDate string) (output *domain.AuthorEntity, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.create", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Create(firstName, lastName, displayName, birthDate)
	return
}

func (mw InstrumentingAuthorMiddleware) List(pageToken, pageSize string, filterParams util.FilterParams) (output []*domain.AuthorEntity, nextToken string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.list", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, nextToken, err = mw.Next.List(pageToken, pageSize, filterParams)
	return
}

func (mw InstrumentingAuthorMiddleware) Get(id string) (output *domain.AuthorEntity, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.get", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Get(id)
	return
}

func (mw InstrumentingAuthorMiddleware) Update(id, firstName, lastName, displayName, birthDate string) (output *domain.AuthorEntity, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.update", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Update(id, firstName, lastName, displayName, birthDate)
	return
}

func (mw InstrumentingAuthorMiddleware) Delete(id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "author.delete", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Delete(id)
	return
}

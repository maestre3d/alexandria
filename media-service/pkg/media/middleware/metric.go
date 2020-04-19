package middleware

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
)

type InstrumentingMediaMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	Next           service.IMediaService
}

func (mw InstrumentingMediaMiddleware) Create(title, displayName, description, userID, authorID, publishDate, mediaType string) (output *domain.MediaEntity, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.create", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Create(title, displayName, description, userID, authorID, publishDate, mediaType)
	return
}

func (mw InstrumentingMediaMiddleware) List(pageToken, pageSize string, filterParams util.FilterParams) (output []*domain.MediaEntity, nextToken string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.list", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, nextToken, err = mw.Next.List(pageToken, pageSize, filterParams)
	return
}

func (mw InstrumentingMediaMiddleware) Get(id string) (output *domain.MediaEntity, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.get", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Get(id)
	return
}

func (mw InstrumentingMediaMiddleware) Update(id, title, displayName, description, userID, authorID, publishDate, mediaType string) (output *domain.MediaEntity, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.update", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	output, err = mw.Next.Update(id, title, displayName, description, userID, authorID, publishDate, mediaType)
	return
}

func (mw InstrumentingMediaMiddleware) Delete(id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "media.delete", "error", fmt.Sprint(err != nil)}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err = mw.Next.Delete(id)
	return
}

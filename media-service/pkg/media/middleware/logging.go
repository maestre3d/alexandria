package middleware

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/media/domain"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/media-service/pkg/media/service"
)

type LoggingMediaMiddleware struct {
	Logger log.Logger
	Next   service.IMediaService
}

func (mw LoggingMediaMiddleware) Create(title, displayName, description, userID, authorID, publishDate, mediaType string) (output *domain.MediaEntity, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.create",
			"input", fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s", title, displayName, userID, authorID, publishDate, mediaType),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Create(title, displayName, description, userID, authorID, publishDate, mediaType)
	return
}

func (mw LoggingMediaMiddleware) List(pageToken, pageSize string, filterParams util.FilterParams) (output []*domain.MediaEntity, nextToken string, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.list",
			"input", fmt.Sprintf("%s, %s, %s", pageToken, pageSize, filterParams),
			"output", fmt.Sprintf("%+v", output),
			"next_token", nextToken,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, nextToken, err = mw.Next.List(pageToken, pageSize, filterParams)
	return
}

func (mw LoggingMediaMiddleware) Get(id string) (output *domain.MediaEntity, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.get",
			"input", fmt.Sprintf("%s", id),
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(id)
	return
}

func (mw LoggingMediaMiddleware) Update(id, title, displayName, description, userID, authorID, publishDate, mediaType string) (output *domain.MediaEntity, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.update",
			"input", fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s", title, displayName, userID, authorID, publishDate, mediaType),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Update(id, title, displayName, description, userID, authorID, publishDate, mediaType)
	return
}

func (mw LoggingMediaMiddleware) Delete(id string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.delete",
			"input", fmt.Sprintf("%s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Delete(id)
	return
}

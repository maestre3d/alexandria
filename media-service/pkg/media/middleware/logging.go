package middleware

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	"time"
)

type LoggingMediaMiddleware struct {
	Logger log.Logger
	Next   usecase.MediaInteractor
}

func (mw LoggingMediaMiddleware) Create(ctx context.Context, aggregate *domain.MediaAggregate) (output *domain.Media, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.create",
			"input", fmt.Sprintf("%+v", aggregate),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Create(ctx, aggregate)
	return
}

func (mw LoggingMediaMiddleware) List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) (output []*domain.Media, nextToken string, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.list",
			"input", fmt.Sprintf("page_token: %s, page_size: %s, filter: %s", pageToken, pageSize, filterParams),
			"output", fmt.Sprintf("%+v", output),
			"next_token", nextToken,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, nextToken, err = mw.Next.List(ctx, pageToken, pageSize, filterParams)
	return
}

func (mw LoggingMediaMiddleware) Get(ctx context.Context, id string) (output *domain.Media, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.get",
			"input", fmt.Sprintf("id: %s", id),
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

func (mw LoggingMediaMiddleware) Update(ctx context.Context, aggregate *domain.MediaUpdateAggregate) (output *domain.Media, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.update",
			"input", fmt.Sprintf("%+v", aggregate),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Update(ctx, aggregate)
	return
}

func (mw LoggingMediaMiddleware) Delete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.delete",
			"input", fmt.Sprintf("id: %s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Delete(ctx, id)
	return
}

func (mw LoggingMediaMiddleware) Restore(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.restore",
			"input", fmt.Sprintf("id: %s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Restore(ctx, id)
	return
}

func (mw LoggingMediaMiddleware) HardDelete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.hard_delete",
			"input", fmt.Sprintf("id: %s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.HardDelete(ctx, id)
	return
}

type LoggingMediaSAGAMiddleware struct {
	Logger log.Logger
	Next   usecase.MediaSAGAInteractor
}

func (mw LoggingMediaSAGAMiddleware) VerifyAuthor(ctx context.Context, rootID string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.saga.verify_author",
			"input", fmt.Sprintf("root_id: %s", rootID),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.VerifyAuthor(ctx, rootID)
	return
}

func (mw LoggingMediaSAGAMiddleware) Done(ctx context.Context, rootID, operation string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.saga.done",
			"input", fmt.Sprintf("root_id: %s, operation: %s", rootID, operation),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Done(ctx, rootID, operation)
	return
}

func (mw LoggingMediaSAGAMiddleware) Failed(ctx context.Context, rootID, operation, backup string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "media.saga.failed",
			"input", fmt.Sprintf("root_id: %s, operation: %s, backup: %s", rootID, operation, backup),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Failed(ctx, rootID, operation, backup)
	return
}

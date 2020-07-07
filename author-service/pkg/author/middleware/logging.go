package middleware

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"time"
)

type LoggingAuthorMiddleware struct {
	Logger log.Logger
	Next   usecase.AuthorInteractor
}

func (mw LoggingAuthorMiddleware) Create(ctx context.Context, aggregate *domain.AuthorAggregate) (output *domain.Author, err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.create",
			"input", fmt.Sprintf("%+v", aggregate),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Create(ctx, aggregate)
	return
}

func (mw LoggingAuthorMiddleware) List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) (output []*domain.Author, nextToken string, err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.list",
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

func (mw LoggingAuthorMiddleware) Get(ctx context.Context, id string) (output *domain.Author, err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.get",
			"input", fmt.Sprintf("id: %s", id),
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

func (mw LoggingAuthorMiddleware) Update(ctx context.Context, aggregate *domain.AuthorUpdateAggregate) (output *domain.Author, err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.update",
			"input", fmt.Sprintf("%+v", aggregate),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Update(ctx, aggregate)
	return
}

func (mw LoggingAuthorMiddleware) Delete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.delete",
			"input", fmt.Sprintf("id: %s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Delete(ctx, id)
	return
}

func (mw LoggingAuthorMiddleware) Restore(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.restore",
			"input", fmt.Sprintf("id: %s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Restore(ctx, id)
	return
}

func (mw LoggingAuthorMiddleware) HardDelete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.hard_delete",
			"input", fmt.Sprintf("id: %s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.HardDelete(ctx, id)
	return
}

type LoggingAuthorSAGAMiddleware struct {
	Logger log.Logger
	Next   usecase.AuthorSAGAInteractor
}

func (mw LoggingAuthorSAGAMiddleware) Verify(ctx context.Context, authorsJSON []byte) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.saga.verify",
			"input", fmt.Sprintf("author_pool: %s", string(authorsJSON)),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Verify(ctx, authorsJSON)
	return
}

func (mw LoggingAuthorSAGAMiddleware) Done(ctx context.Context, rootID, operation string) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.saga.done",
			"input", fmt.Sprintf("root_id: %s, operation: %s", rootID, operation),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Done(ctx, rootID, operation)
	return
}

func (mw LoggingAuthorSAGAMiddleware) Failed(ctx context.Context, rootID, operation, snapshot string) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.saga.failed",
			"input", fmt.Sprintf("root_id: %s, operation: %s, snapshot: %s", rootID, operation, snapshot),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Failed(ctx, rootID, operation, snapshot)
	return
}

func (mw LoggingAuthorSAGAMiddleware) UpdatePicture(ctx context.Context, rootID string, urlJSON []byte) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.saga.update_picture",
			"input", fmt.Sprintf("root_id: %s, url: %s", rootID, string(urlJSON)),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.UpdatePicture(ctx, rootID, urlJSON)
	return
}

func (mw LoggingAuthorSAGAMiddleware) RemovePicture(ctx context.Context, rootID string) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "author.saga.remove_picture",
			"input", fmt.Sprintf("root_id: %s", rootID),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.RemovePicture(ctx, rootID)
	return
}

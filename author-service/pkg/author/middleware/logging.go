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
		mw.Logger.Log(
			"method", "author.create",
			"input", fmt.Sprintf("%s, %s, %s, %s, %s", aggregate.FirstName, aggregate.LastName, aggregate.DisplayName,
				aggregate.BirthDate, aggregate.OwnerID),
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
		mw.Logger.Log(
			"method", "author.list",
			"input", fmt.Sprintf("%s, %s, %s", pageToken, pageSize, filterParams),
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
		mw.Logger.Log(
			"method", "author.get",
			"input", fmt.Sprintf("%s", id),
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

func (mw LoggingAuthorMiddleware) Update(ctx context.Context, id, status string, aggregate *domain.AuthorAggregate) (output *domain.Author, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "author.update",
			"input", fmt.Sprintf("%s, %s, %s, %s, %s, %s", id, aggregate.FirstName, aggregate.LastName,
				aggregate.DisplayName, aggregate.BirthDate, aggregate.OwnerID),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Update(ctx, id, status, aggregate)
	return
}

func (mw LoggingAuthorMiddleware) Delete(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "author.delete",
			"input", fmt.Sprintf("%s", id),
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
		mw.Logger.Log(
			"method", "author.restore",
			"input", fmt.Sprintf("%s", id),
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
		mw.Logger.Log(
			"method", "author.hard_delete",
			"input", fmt.Sprintf("%s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Delete(ctx, id)
	return
}

package middleware

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/author-service/pkg/author/service"
	"time"
)

type LoggingAuthorMiddleware struct {
	Logger log.Logger
	Next service.IAuthorService
}

func (mw LoggingAuthorMiddleware) Create(firstName, lastName, displayName, birthDate string) (output *domain.AuthorEntity, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "author.create",
			"input", fmt.Sprintf("%s, %s, %s, %s", firstName, lastName, displayName, birthDate),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Create(firstName, lastName, displayName, birthDate)
	return
}

func (mw LoggingAuthorMiddleware) List(pageToken, pageSize string, filterParams util.FilterParams) (output []*domain.AuthorEntity, nextToken string, err error) {
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

	output, nextToken, err = mw.Next.List(pageToken, pageSize, filterParams)
	return
}

func (mw LoggingAuthorMiddleware) Get(id string) (output *domain.AuthorEntity, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "author.get",
			"input", fmt.Sprintf("%s", id),
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(id)
	return
}

func (mw LoggingAuthorMiddleware) Update(id, firstName, lastName, displayName, birthDate string) (output *domain.AuthorEntity, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "author.update",
			"input", fmt.Sprintf("%s, %s, %s, %s, %s", id, firstName, lastName, displayName, birthDate),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Update(id, firstName, lastName, displayName, birthDate)
	return
}

func (mw LoggingAuthorMiddleware) Delete(id string) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "author.delete",
			"input", fmt.Sprintf("%s", id),
			"output", err,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Delete(id)
	return
}

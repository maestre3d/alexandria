package middleware

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	"time"
)

type LoggingBlobMiddleware struct {
	Logger log.Logger
	Next   usecase.BlobInteractor
}

func (mw LoggingBlobMiddleware) Store(ctx context.Context, aggregate *domain.BlobAggregate) (output *domain.Blob, err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "blob.store",
			"input", fmt.Sprintf("%+v", aggregate),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Store(ctx, aggregate)
	return
}

func (mw LoggingBlobMiddleware) Get(ctx context.Context, id, service string) (output *domain.Blob, err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "blob.get",
			"input", fmt.Sprintf("root_id: %s, service: %s", id, service),
			"output", fmt.Sprintf("%+v", output),
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(ctx, id, service)
	return
}

func (mw LoggingBlobMiddleware) Delete(ctx context.Context, id, service string) (err error) {
	defer func(begin time.Time) {
		_ = mw.Logger.Log(
			"method", "blob.delete",
			"input", fmt.Sprintf("root_id: %s, service: %s", id, service),
			"output", "",
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Delete(ctx, id, service)
	return
}

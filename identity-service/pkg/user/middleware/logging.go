package middleware

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
	"time"
)

type LoggingUserMiddleware struct {
	Logger log.Logger
	Next   usecase.UserInteractor
}

func (mw LoggingUserMiddleware) Get(ctx context.Context, id string) (output *domain.User, err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "user.get",
			"input", fmt.Sprintf("id: %s", id),
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

type LoggingUserSAGAMiddleware struct {
	Logger log.Logger
	Next   usecase.UserSAGAInteractor
}

func (mw LoggingUserSAGAMiddleware) Verify(ctx context.Context, service string, usersJSON []byte) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "user.saga.verify",
			"input", fmt.Sprintf("service: %s, user_pool: %s", service, string(usersJSON)),
			"output", "",
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.Verify(ctx, service, usersJSON)
	return
}

func (mw LoggingUserSAGAMiddleware) UpdatePicture(ctx context.Context, id string, urlJSON []byte) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "user.saga.update_picture",
			"input", fmt.Sprintf("user_id: %s, picture_url: %s", id, string(urlJSON)),
			"output", "",
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.UpdatePicture(ctx, id, urlJSON)
	return
}

func (mw LoggingUserSAGAMiddleware) RemovePicture(ctx context.Context, rootJSON []byte) (err error) {
	defer func(begin time.Time) {
		mw.Logger.Log(
			"method", "user.saga.update_picture",
			"input", fmt.Sprintf("user_id: %s", string(rootJSON)),
			"output", "",
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.Next.RemovePicture(ctx, rootJSON)
	return
}

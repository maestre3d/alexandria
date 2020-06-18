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
			"input", fmt.Sprintf("%s", id),
			"output", output,
			"err", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	output, err = mw.Next.Get(ctx, id)
	return
}

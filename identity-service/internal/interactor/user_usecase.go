package interactor

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
)

type UserUseCase struct {
	log        log.Logger
	repository domain.UserRepository
}

func NewUserUseCase(logger log.Logger, repo domain.UserRepository) *UserUseCase {
	return &UserUseCase{
		log:        logger,
		repository: repo,
	}
}

func (u *UserUseCase) Get(ctx context.Context, id string) (*domain.User, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	return u.repository.FetchByID(ctxR, id)
}
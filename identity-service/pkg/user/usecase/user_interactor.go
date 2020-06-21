package usecase

import (
	"context"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
)

type UserInteractor interface {
	Get(ctx context.Context, id string) (*domain.User, error)
}

type UserSAGAInteractor interface {
	Verify(ctx context.Context, usersJSON []byte) error
}

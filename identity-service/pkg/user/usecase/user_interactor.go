package usecase

import (
	"context"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
)

type UserInteractor interface {
	Get(ctx context.Context, id string) (*domain.User, error)
}

type UserSAGAInteractor interface {
	Verify(ctx context.Context, service string, usersJSON []byte) error
	UpdatePicture(ctx context.Context, id string, urlJSON []byte) error
	RemovePicture(ctx context.Context, rootJSON []byte) error
}

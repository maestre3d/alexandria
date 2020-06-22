package interactor

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
)

type User struct {
	logger     log.Logger
	repository domain.UserRepository
}

func NewUser(logger log.Logger, repo domain.UserRepository) *User {
	return &User{
		logger:     logger,
		repository: repo,
	}
}

// TODO: Add the following use cases
// Read
// - List
// Write
// - Update
// - Delete (deactivate/soft-delete)
// - Restore (activate)
// - HardDelete (hard-delete)

func (u *User) Get(ctx context.Context, id string) (*domain.User, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()
	return u.repository.FetchByID(ctxR, id)
}

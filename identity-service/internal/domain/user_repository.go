package domain

import (
	"context"
	"github.com/alexandria-oss/core"
)

type UserRepository interface {
	// Save store a user entity
	Save(ctx context.Context, user User) error
	// Fetch lists a set of user's entities using pagination tokens and filtering
	Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*User, error)
	// FetchOne get a user entity by it's username or external_id
	FetchOne(ctx context.Context, token string) (*User, error)
	// Replace update a user atomically
	Replace(ctx context.Context, user User) error
	// Remove softly remove a user from storage
	Remove(ctx context.Context, token string) error
	// Restore get back a user previously removed softly
	Restore(ctx context.Context, token string) error
	// HardRemove delete a user permanently from storage
	HardRemove(ctx context.Context, token string) error
}

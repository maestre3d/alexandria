package domain

import (
	"context"
	"github.com/alexandria-oss/core"
)

// AuthorRepository Author entity's repository
type AuthorRepository interface {
	Save(ctx context.Context, author Author) error
	SaveRaw(ctx context.Context, author Author) error
	Fetch(ctx context.Context, params core.PaginationParams, filterParams core.FilterParams) ([]*Author, error)
	FetchByID(ctx context.Context, id string, showDisabled bool) (*Author, error)
	Replace(ctx context.Context, author Author) error
	Remove(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	HardRemove(ctx context.Context, id string) error
	ChangeState(ctx context.Context, id, state string) error
}

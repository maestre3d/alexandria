package domain

import (
	"context"
	"github.com/alexandria-oss/core"
)

type CategoryRepository interface {
	Save(ctx context.Context, category Category) error
	Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*Category, error)
	FetchByID(ctx context.Context, id string, activeOnly bool) (*Category, error)
	Replace(ctx context.Context, category Category) error
	Remove(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	HardRemove(ctx context.Context, id string) error
}

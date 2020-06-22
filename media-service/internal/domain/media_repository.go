package domain

import (
	"context"
	"github.com/alexandria-oss/core"
)

type MediaRepository interface {
	Save(ctx context.Context, media Media) error
	SaveRaw(ctx context.Context, media Media) error
	Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*Media, error)
	FetchByID(ctx context.Context, id string, showDisabled bool) (*Media, error)
	Replace(ctx context.Context, media Media) error
	Remove(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	HardRemove(ctx context.Context, id string) error
	ChangeState(ctx context.Context, id, state string) error
}

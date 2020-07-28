package domain

import (
	"context"
	"github.com/alexandria-oss/core"
)

type CategoryRootRepository interface {
	Save(ctx context.Context, root CategoryByRoot) error
	AddItem(ctx context.Context, rootID string, item map[string]string) error
	FetchByRoot(ctx context.Context, rootID string) (*CategoryByRoot, error)
	Fetch(ctx context.Context, params core.PaginationParams) ([]*CategoryByRoot, error)
	RemoveItem(ctx context.Context, rootID, categoryID string) error
	HardRemoveList(ctx context.Context, rootID string) error
}

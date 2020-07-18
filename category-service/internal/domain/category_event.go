package domain

import (
	"context"
)

const (
	// Domain events - Side-effects
	CategoryCreated     string = "CATEGORY_CREATED"             // Produced
	CategoryUpdated     string = "CATEGORY_UPDATED"             // Produced
	CategoryRemoved     string = "CATEGORY_REMOVED"             // Produced
	CategoryHardRemoved string = "CATEGORY_PERMANENTLY_REMOVED" // Produced
)

type CategoryEventBus interface {
	Created(ctx context.Context, category Category) error
	Updated(ctx context.Context, category Category) error
	Removed(ctx context.Context, id string) error
	HardRemoved(ctx context.Context, id string) error
}

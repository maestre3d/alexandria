package domain

import "context"

const (
	// Integration events

	// Domain events
	CategoryRootCreated     = "CATEGORY_ROOT_CREATED"             // Produced
	CategoryRootHardRemoved = "CATEGORY_ROOT_PERMANENTLY_REMOVED" // Produced
)

type CategoryRootEventBus interface {
	StartCreate(ctx context.Context, root CategoryByRoot) error
	Created(ctx context.Context, root CategoryByRoot) error
	HardRemoved(ctx context.Context, id string) error
}

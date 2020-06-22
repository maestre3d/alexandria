package domain

import "context"

const (
	AuthorUpdated     = "AUTHOR_UPDATED"
	AuthorRemoved     = "AUTHOR_REMOVED"
	AuthorRestored    = "AUTHOR_RESTORED"
	AuthorHardRemoved = "AUTHOR_PERMANENTLY_REMOVED"
)

type AuthorEventBus interface {
	StartCreate(ctx context.Context, author Author) error
	StartUpdate(ctx context.Context, author Author, backup Author) error
	Updated(ctx context.Context, author Author) error
	Removed(ctx context.Context, id string) error
	Restored(ctx context.Context, id string) error
	HardRemoved(ctx context.Context, id string) error
}

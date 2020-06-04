package domain

import "context"

const (
	AuthorCreated     = "ALEXANDRIA_AUTHOR_CREATED"
	AuthorDeleted     = "ALEXANDRIA_AUTHOR_DELETED"
	AuthorHardDeleted = "ALEXANDRIA_AUTHOR_HARD_DELETED"
	AuthorRestored    = "ALEXANDRIA_AUTHOR_RESTORED"
)

type IAuthorEventBus interface {
	Created(ctx context.Context, author *Author) error
	Deleted(ctx context.Context, id string, isHard bool) error
	Restored(ctx context.Context, id string) error
}

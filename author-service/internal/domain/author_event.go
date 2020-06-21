package domain

import "context"

const (
	AuthorCreated     = "AUTHOR_CREATED"
	AuthorUpdated     = "AUTHOR_UPDATED"
	AuthorDeleted     = "AUTHOR_DELETED"
	AuthorHardDeleted = "AUTHOR_PERMANENTLY_DELETED"
	AuthorRestored    = "AUTHOR_RESTORED"
	OwnerVerify       = "OWNER_VERIFY"
	OwnerVerified     = "AUTHOR_OWNER_VERIFIED"
	OwnerFailed       = "AUTHOR_OWNER_FAILED"
)

type AuthorEventBus interface {
	StartCreate(ctx context.Context, author *Author) error
	Created(ctx context.Context, author *Author) error
	StartUpdate(ctx context.Context, author *Author, backup *Author) error
	Updated(ctx context.Context, author *Author) error
	Deleted(ctx context.Context, id string, isHard bool) error
	Restored(ctx context.Context, id string) error
}

package domain

import "context"

const (
	MediaCreated        = "MEDIA_CREATED"
	MediaUpdated        = "MEDIA_UPDATED"
	MediaDeleted        = "MEDIA_REMOVED"
	MediaRestored       = "MEDIA_RESTORED"
	MediaHardDeleted    = "MEDIA_PERMANENTLY_REMOVED"
	OwnerVerify         = "OWNER_VERIFY"
	MediaOwnerVerified  = "MEDIA_OWNER_VERIFIED"
	MediaOwnerFailed    = "MEDIA_OWNER_FAILED"
	AuthorVerify        = "AUTHOR_VERIFY"
	MediaAuthorVerified = "MEDIA_AUTHOR_VERIFIED"
	MediaAuthorFailed   = "MEDIA_AUTHOR_FAILED"
)

type MediaEvent interface {
	StartCreate(ctx context.Context, media Media) error
	Created(ctx context.Context, media Media) error
	StartUpdate(ctx context.Context, media Media, backup Media) error
	Updated(ctx context.Context, media Media) error
	Removed(ctx context.Context, id string) error
	Restored(ctx context.Context, id string) error
	HardRemoved(ctx context.Context, id string) error
}

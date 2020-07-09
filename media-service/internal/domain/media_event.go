package domain

import "context"

const (
	MediaUpdated     = "MEDIA_UPDATED"
	MediaRemoved     = "MEDIA_REMOVED"
	MediaRestored    = "MEDIA_RESTORED"
	MediaHardRemoved = "MEDIA_PERMANENTLY_REMOVED"
)

type MediaEvent interface {
	StartCreate(ctx context.Context, media Media) error
	StartUpdate(ctx context.Context, media Media, snapshot Media) error
	Updated(ctx context.Context, media Media) error
	Removed(ctx context.Context, id string) error
	Restored(ctx context.Context, id string) error
	HardRemoved(ctx context.Context, id string) error
}

package domain

import (
	"context"
)

const (
	// Side-effect events
	MediaCreated = "MEDIA_CREATED" // Produced

	// Foreign validation events
	OwnerVerify    = "OWNER_VERIFY"          // Produced
	OwnerVerified  = "MEDIA_OWNER_VERIFIED"  // Consumed
	OwnerFailed    = "MEDIA_OWNER_FAILED"    // Consumed
	AuthorVerify   = "AUTHOR_VERIFY"         // Produced
	AuthorVerified = "MEDIA_AUTHOR_VERIFIED" // Consumed
	AuthorFailed   = "MEDIA_AUTHOR_FAILED"   // Consumed
	BlobUploaded   = "MEDIA_BLOB_UPLOADED"   // Consumed
	BlobRemoved    = "MEDIA_BLOB_REMOVED"    // Consumed
	BlobFailed     = "BLOB_FAILED"           // Produced
)

type MediaEventSAGA interface {
	VerifyAuthor(ctx context.Context, authors []string) error
	Created(ctx context.Context, media Media) error
	BlobFailed(ctx context.Context, msg string) error
}

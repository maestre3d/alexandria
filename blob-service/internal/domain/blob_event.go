package domain

import "context"

const (
	// Start Transaction
	BlobUploaded = "BLOB_UPLOADED" // Produced

	// Side-effect events
	BlobRemoved = "BLOB_REMOVED" // Produced

	// Foreign validation events
	BlobFailed = "BLOB_FAILED" // Consumed
)

type BlobEvent interface {
	Uploaded(ctx context.Context, blob Blob) error
	Removed(ctx context.Context, rootID, service string) error
}

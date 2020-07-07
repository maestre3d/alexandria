package domain

import "context"

const (
	// Start Transaction
	BlobUploaded = "BLOB_UPLOADED" // Produced

	// Side-effect events
	BlobRemoved = "BLOB_REMOVED" // Produced
)

type BlobEvent interface {
	Uploaded(ctx context.Context, blob Blob, snapshot *Blob) error
	Removed(ctx context.Context, rootID, service string) error
}

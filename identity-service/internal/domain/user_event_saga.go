package domain

import "context"

const (
	// Bounded context validation events
	OwnerVerify   = "OWNER_VERIFY"   // Consumed
	OwnerVerified = "OWNER_VERIFIED" // Produced (service_name+"_"+event)
	OwnerFailed   = "OWNER_FAILED"   // Produced (service_name+"_"+event)

	// External helper services
	BlobUploaded = "USER_BLOB_UPLOADED" // Consumed
	BlobRemoved  = "USER_BLOB_REMOVED"  // Consumed
	BlobFailed   = "BLOB_FAILED"        // Produced
)

type UserEventSAGA interface {
	Verified(ctx context.Context, service string) error
	Failed(ctx context.Context, service, msg string) error
	BlobFailed(ctx context.Context, msg string) error
}

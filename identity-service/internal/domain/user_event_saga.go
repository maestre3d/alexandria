package domain

import "context"

const (
	// Bounded context validation events
	OwnerVerify   = "OWNER_VERIFY"   // Consumed
	OwnerVerified = "OWNER_VERIFIED" // Produced (service_name+"_"+event)
	OwnerFailed   = "OWNER_FAILED"   // Produced (service_name+"_"+event)
)

type UserEventSAGA interface {
	OwnerVerified(ctx context.Context) error
	OwnerFailed(ctx context.Context) error
}

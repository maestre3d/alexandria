package domain

import "context"

const (
	// Bounded context validation events
	OwnerVerify   = "OWNER_VERIFY"   // Consumed
	OwnerVerified = "OWNER_VERIFIED" // Produced (service_name+"_"+event)
	OwnerFailed   = "OWNER_FAILED"   // Produced (service_name+"_"+event)
)

type UserEventSAGA interface {
	Verified(ctx context.Context) error
	Failed(ctx context.Context) error
}

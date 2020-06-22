package domain

import "context"

const (
	// Side-effect events
	AuthorCreated = "AUTHOR_CREATED" // Produced

	// Foreign validation events
	OwnerVerify   = "OWNER_VERIFY"          // Produced
	OwnerVerified = "AUTHOR_OWNER_VERIFIED" // Consumed
	OwnerFailed   = "AUTHOR_OWNER_FAILED"   // Consumed

	// Bounded context validation events
	AuthorVerify   = "AUTHOR_VERIFY"   // Consumed
	AuthorVerified = "AUTHOR_VERIFIED" // Produced (service_name+"_"+event)
	AuthorFailed   = "AUTHOR_FAILED"   // Produced (service_name+"_"+event)
)

type AuthorSAGAEventBus interface {
	Verified(ctx context.Context) error
	Failed(ctx context.Context) error
	Created(ctx context.Context, author Author) error
}

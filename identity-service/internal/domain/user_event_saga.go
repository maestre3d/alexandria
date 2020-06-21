package domain

import "context"

const (
	OwnerVerify   = "OWNER_VERIFY"
	OwnerVerified = "OWNER_VERIFIED"
	OwnerFailed   = "OWNER_FAILED"
)

type UserEventSAGA interface {
	OwnerVerified(ctx context.Context) error
	OwnerFailed(ctx context.Context) error
}

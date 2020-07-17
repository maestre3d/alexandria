package domain

import (
	"context"
	"fmt"
	"strings"
)

// Integration events - Transactions used as heterogeneous entity

// Produced
func EntityVerify(service string) string {
	return fmt.Sprintf("%s_VERIFY", strings.ToUpper(service))
}

// Consumed
func EntityVerified(service string) string {
	return fmt.Sprintf("CATEGORY_%s_VERIFIED", strings.ToUpper(service))
}

// Consumed
func EntityFailed(service string) string {
	return fmt.Sprintf("CATEGORY_%s_FAILED", strings.ToUpper(service))
}

type CategorySAGAEventBus interface {
	Created(ctx context.Context, category Category) error
	Failed(ctx context.Context, id string) error
}

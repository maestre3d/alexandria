package core

import "github.com/google/uuid"

// ValidateUUID Validate a Unique Universal ID (UUID)
func ValidateUUID(id string) error {
	_, err := uuid.Parse(id)
	return err
}

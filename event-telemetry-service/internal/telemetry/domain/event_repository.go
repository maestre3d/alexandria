package domain

import (
	"github.com/alexandria-oss/core"
)

// EventRepository Store manager for telemetry entities
type EventRepository interface {
	// Save Store a event entity
	Save(event *EventEntity) error
	// Fetch Get event entities
	Fetch(params *core.PaginationParams, filterParams core.FilterParams) ([]*EventEntity, error)
	// FetchByID Get an specific event entity
	FetchByID(id string) (*EventEntity, error)
	// Update Update an specific event entity
	Update(eventUpdated *EventEntity) error
	// Remove Delete an specific event entity (hard-delete)
	Remove(id string) error
}

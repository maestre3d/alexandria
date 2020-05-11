package domain

import (
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/eventbus"
)

// EventRepository Store manager for telemetry entities
type EventRepository interface {
	// Save Store a event entity
	Save(event *eventbus.Event) error
	// Fetch Get event entities
	Fetch(params *core.PaginationParams, filterParams core.FilterParams) ([]*eventbus.Event, error)
	// FetchByID Get an specific event entity
	FetchByID(id string) (*eventbus.Event, error)
	// Update Update an specific event entity
	Update(eventUpdated *eventbus.Event) error
	// Remove Delete an specific event entity (hard-delete)
	Remove(id string) error
}

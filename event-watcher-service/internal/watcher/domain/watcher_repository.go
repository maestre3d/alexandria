package domain

import "github.com/maestre3d/alexandria/event-watcher-service/internal/shared/domain/util"

// WatcherRepository Store manager for watcher entities
type WatcherRepository interface {
	// Save Store a watcher entity
	Save(watcher *WatcherEntity) error
	// Fetch Get watcher entities
	Fetch(params *util.PaginationParams, filterParams util.FilterParams) ([]*WatcherEntity, error)
	// FetchByID Get an specific watcher entity
	FetchByID(id string) (*WatcherEntity, error)
	// Update Update an specific watcher entity
	Update(watcherUpdated *WatcherEntity) error
	// Remove Delete an specific watcher entity (hard-delete)
	Remove(id string) error
}

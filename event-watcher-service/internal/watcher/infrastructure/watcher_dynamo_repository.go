package infrastructure

import (
	"context"
	"io"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/domain"
	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/awsdynamodb"
)

// WatcherDynamoRepository Watcher Repository interface implementation in AWS DynamoDB
type WatcherDynamoRepository struct {
	ctx    context.Context
	mtx    *sync.Mutex
	logger log.Logger
	db     *docstore.Collection
}

func NewWatcherDynamoRepository(ctx context.Context, logger log.Logger, coll *docstore.Collection) *WatcherDynamoRepository {
	return &WatcherDynamoRepository{ctx, new(sync.Mutex), logger, coll}
}

func (r *WatcherDynamoRepository) Save(watcher *domain.WatcherEntity) error {
	// Lock struct's mutability
	r.mtx.Lock()
	defer r.mtx.Unlock()

	err := r.db.Create(r.ctx, watcher)
	if err != nil {
		return err
	}

	return nil
}

func (r *WatcherDynamoRepository) Fetch(params *util.PaginationParams, filterParams util.FilterParams) ([]*domain.WatcherEntity, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Get watcher with params token, this is needed to simulate SQL's subquery/join to make use of entity's timestamp
	watcherToken, err := r.FetchByID(params.Token)
	if err != nil {
		return nil, err
	}

	// Atomic querying - Fluent API-like
	query := r.db.Query()
	for filter, value := range filterParams {
		switch {
		case filter == "query" && value != "":
			QueryCriteriaDynamo(value, query)
		}
	}

	// Default query (footer)
	iter := query.Where("watcher_id", "=", watcherToken.ID).Where("dispatch_time", ">=", watcherToken.DispatchTime).
		OrderBy("dispatch_time", docstore.Ascending).Limit(params.Size).Get(r.ctx)
	defer iter.Stop()

	watchers := make([]*domain.WatcherEntity, 0)
	for {
		watcher := new(domain.WatcherEntity)

		err := iter.Next(r.ctx, watcher)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else {
			watchers = append(watchers, watcher)
		}
	}

	return watchers, nil
}

func (r *WatcherDynamoRepository) FetchByID(id string) (*domain.WatcherEntity, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	watcher := new(domain.WatcherEntity)
	watcher.ID = id

	err := r.db.Get(r.ctx, watcher, "watcher_id")
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func (r *WatcherDynamoRepository) Update(watcherUpdated *domain.WatcherEntity) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.db.Actions().Replace(watcherUpdated).Do(r.ctx)
}

func (r *WatcherDynamoRepository) Remove(id string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	watcher := new(domain.WatcherEntity)
	watcher.ID = id

	return r.db.Actions().Get(watcher, "watcher_id").Delete(watcher).Do(r.ctx)
}

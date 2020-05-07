package infrastructure

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/event-telemetry-service/internal/telemetry/domain"
	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/awsdynamodb"
	"gocloud.dev/gcerrors"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EventDynamoRepository Watcher Repository interface implementation in AWS DynamoDB
type EventDynamoRepository struct {
	ctx    context.Context
	mtx    *sync.Mutex
	logger log.Logger
	db     *docstore.Collection
}

type eventDocEntity struct {
	ID               string `json:"event_id" docstore:"event_id"`
	ServiceName      string `json:"service_name" docstore:"service_name"`
	TransactionID    string `json:"transaction_id,omitempty" docstore:"transaction_id,omitempty"`
	EventType        string `json:"event_type" docstore:"event_type"`
	Content          string `json:"content" docstore:"content"`
	Importance       string `json:"importance" docstore:"importance"`
	Provider         string `json:"provider" docstore:"provider"`
	DispatchTime     int64  `json:"dispatch_time" docstore:"dispatch_time"`
	DocstoreRevision interface{}
}

func NewEventDynamoRepository(ctx context.Context, logger log.Logger, coll *docstore.Collection) *EventDynamoRepository {
	return &EventDynamoRepository{ctx, new(sync.Mutex), logger, coll}
}

func (r *EventDynamoRepository) Save(event *domain.EventEntity) error {
	// Lock struct's mutability
	r.mtx.Lock()
	defer r.mtx.Unlock()

	err := r.db.Create(r.ctx, event)
	if err != nil {
		return err
	}

	return nil
}

func (r *EventDynamoRepository) Fetch(params *core.PaginationParams, filterParams core.FilterParams) ([]*domain.EventEntity, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Atomic querying - Fluent API-like
	query := r.db.Query()
	for filter, value := range filterParams {
		switch {
		case filter == "query" && value != "":
			QueryCriteriaDynamo(value, query)
		}
	}

	// Get telemetry with params token, this is needed to simulate SQL's sub-query/join to make use of entity's timestamp
	// Get sort_key
	dispatch, err := strconv.ParseInt(params.Token, 10, 64)
	if err != nil {
		dispatch = time.Now().UnixNano() / 1000000
	}

	// Default query (footer)
	iter := query.Where("dispatch_time", "<=", dispatch).Limit(params.Size).Get(r.ctx)
	defer iter.Stop()

	events := make([]*domain.EventEntity, 0)
	for {
		event := new(domain.EventEntity)
		err := iter.Next(r.ctx, event)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else {
			events = append(events, event)
		}
	}

	// Order by timestamp manually - ASC
	for i := range events {
		aux := events[i]
		for y := i; y > 0; y-- {
			if events[y].DispatchTime > events[y-1].DispatchTime {
				events[y] = events[y-1]
				events[y-1] = aux
			} else {
				break
			}
		}
	}

	// Attach timestamp (sort-key) to nextID token
	if len(events) >= params.Size {
		events[len(events)-1].ID = strconv.FormatInt(events[len(events)-1].DispatchTime, 10)
	}

	return events, nil
}

func (r *EventDynamoRepository) FetchByID(id string) (*domain.EventEntity, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Separate partition and sort keys (NoSQL)
	// Keys are separated by '#'
	keys := strings.Split(id, "#")
	if len(keys) > 1 {
		if err := core.ValidateUUID(keys[0]); err != nil {
			return nil, exception.InvalidID
		}
	} else if len(keys) <= 1 {
		return nil, exception.InvalidID
	}

	dispatch, err := strconv.ParseInt(keys[1], 10, 64)
	if err != nil {
		dispatch = time.Now().UnixNano() / 1000000
	}

	eventDoc := &eventDocEntity{ID: keys[0], DispatchTime: dispatch}
	err = r.db.Actions().Get(eventDoc).Do(r.ctx)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			return nil, nil
		}

		return nil, err
	}

	return toEntity(eventDoc), nil
}

func (r *EventDynamoRepository) Update(eventUpdated *domain.EventEntity) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.mtx.Unlock()
	event, err := r.FetchByID(eventUpdated.ID)
	r.mtx.Lock()
	if err != nil {
		return err
	} else if event == nil {
		return exception.EntityNotFound
	}

	eventUpdated.ID = event.ID
	event.DispatchTime = time.Now().UnixNano() / 1000000

	return r.db.Replace(r.ctx, eventUpdated)
}

func (r *EventDynamoRepository) Remove(id string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.mtx.Unlock()
	event, err := r.FetchByID(id)
	r.mtx.Lock()
	if err != nil {
		return err
	} else if event == nil {
		return exception.EntityNotFound
	}

	return r.db.Delete(r.ctx, event)
}

/* Model binding */

/*
func toDocEntity(entity *domain.EventEntity) *eventDocEntity {
	transaction := ""
	if entity.TransactionID != nil {
		transaction = *entity.TransactionID
	}
	return &eventDocEntity{
		ID:               entity.ID,
		ServiceName:      entity.ServiceName,
		TransactionID:    transaction,
		EventType:        entity.EventType,
		Content:          entity.Content,
		Importance:       entity.Importance,
		Provider:         entity.Provider,
		DispatchTime:     strconv.FormatInt(entity.DispatchTime, 10),
		DocstoreRevision: nil,
	}
}*/

func toEntity(docEntity *eventDocEntity) *domain.EventEntity {
	transaction := ""
	if docEntity.TransactionID != "" {
		transaction = docEntity.TransactionID
	}

	return &domain.EventEntity{
		ID:            docEntity.ID,
		ServiceName:   docEntity.ServiceName,
		TransactionID: &transaction,
		EventType:     docEntity.EventType,
		Content:       docEntity.Content,
		Importance:    docEntity.Importance,
		Provider:      docEntity.Provider,
		DispatchTime:  docEntity.DispatchTime,
	}
}

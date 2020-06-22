package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/sony/gobreaker"
	"gocloud.dev/pubsub"
	"sync"
	"time"
)

type AuthorKafkaEventBus struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewAuthorKafkaEventBus(cfg *config.Kernel) *AuthorKafkaEventBus {
	return &AuthorKafkaEventBus{cfg, new(sync.Mutex)}
}

func (b AuthorKafkaEventBus) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "author_kafka_" + action,
		MaxRequests: 1,
		Interval:    0,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: nil,
	}

	return gobreaker.NewCircuitBreaker(st)
}

func (b *AuthorKafkaEventBus) StartCreate(ctx context.Context, author domain.Author) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	ownerPool := make([]string, 0)
	ownerPool = append(ownerPool, author.OwnerID)
	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, ownerJSON)
	t := eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    author.ExternalID,
		SpanID:    "",
		TraceID:   "",
		Operation: domain.AuthorCreated,
	}

	topic, err := eventbus.NewKafkaProducer(ctx, domain.OwnerVerify)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"root_id":        t.RootID,
			"span_id":        t.SpanID,
			"trace_id":       t.TraceID,
			"operation":      t.Operation,
			"service":        e.ServiceName,
			"event_id":       e.ID,
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       e.Provider,
			"dispatch_time":  e.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = b.defaultCircuitBreaker("start_create").Execute(func() (interface{}, error) {
		return nil, topic.Send(ctx, m)
	})

	return err
}

func (b *AuthorKafkaEventBus) StartUpdate(ctx context.Context, author domain.Author, backup domain.Author) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	ownerPool := make([]string, 0)
	ownerPool = append(ownerPool, author.OwnerID)
	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"backup", "author entity"))
	}

	t := &eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    author.ExternalID,
		SpanID:    "",
		TraceID:   "",
		Operation: domain.AuthorUpdated,
		Backup:    string(backupJSON),
	}
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, ownerJSON)

	topic, err := eventbus.NewKafkaProducer(ctx, domain.OwnerVerify)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"root_id":        t.RootID,
			"span_id":        t.SpanID,
			"trace_id":       t.TraceID,
			"operation":      t.Operation,
			"backup":         t.Backup,
			"service":        e.ServiceName,
			"event_id":       e.ID,
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       e.Provider,
			"dispatch_time":  e.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = b.defaultCircuitBreaker("start_update").Execute(func() (interface{}, error) {
		return nil, topic.Send(ctx, m)
	})

	return err
}

func (b *AuthorKafkaEventBus) Updated(ctx context.Context, author domain.Author) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	authorJSON, err := json.Marshal(author)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"author", "author entity"))
	}

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorUpdated)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, authorJSON)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"service":       e.ServiceName,
			"event_id":      e.ID,
			"event_type":    e.EventType,
			"priority":      e.Priority,
			"provider":      e.Provider,
			"dispatch_time": e.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = b.defaultCircuitBreaker("updated").Execute(func() (interface{}, error) {
		return nil, topic.Send(ctx, m)
	})

	return err
}

func (b *AuthorKafkaEventBus) Removed(ctx context.Context, id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Send domain event, Spread side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorRemoved)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	m := &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"service":       e.ServiceName,
			"event_id":      e.ID,
			"event_type":    e.EventType,
			"priority":      e.Priority,
			"provider":      e.Provider,
			"dispatch_time": e.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = b.defaultCircuitBreaker("removed").Execute(func() (interface{}, error) {
		return nil, topic.Send(ctx, m)
	})

	return err
}

func (b *AuthorKafkaEventBus) Restored(ctx context.Context, id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Send domain event, Spread side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorRestored)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	m := &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"service":       e.ServiceName,
			"event_id":      e.ID,
			"event_type":    e.EventType,
			"priority":      e.Priority,
			"provider":      e.Provider,
			"dispatch_time": e.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = b.defaultCircuitBreaker("restored").Execute(func() (interface{}, error) {
		return nil, topic.Send(ctx, m)
	})

	return err
}

func (b *AuthorKafkaEventBus) HardRemoved(ctx context.Context, id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Send domain event, Spread side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorHardRemoved)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	m := &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"service":       e.ServiceName,
			"event_id":      e.ID,
			"event_type":    e.EventType,
			"priority":      e.Priority,
			"provider":      e.Provider,
			"dispatch_time": e.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = b.defaultCircuitBreaker("removed").Execute(func() (interface{}, error) {
		return nil, topic.Send(ctx, m)
	})

	return err
}

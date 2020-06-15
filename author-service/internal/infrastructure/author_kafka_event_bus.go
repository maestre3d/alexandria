package infrastructure

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"gocloud.dev/pubsub"
	"sync"
)

type AuthorKafkaEventBus struct {
	cfg *config.Kernel
	mtx *sync.Mutex
}

func NewAuthorKafkaEventBus(cfg *config.Kernel) *AuthorKafkaEventBus {
	return &AuthorKafkaEventBus{cfg, new(sync.Mutex)}
}

func (b *AuthorKafkaEventBus) StartCreate(ctx context.Context, author *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorPending)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, authorJSON, false)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": uuid.New().String(),
			"operation":      domain.AuthorCreated,
			"service":        e.ServiceName,
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       e.Provider,
		},
		BeforeSend: nil,
	}

	return topic.Send(ctx, m)
}

func (b *AuthorKafkaEventBus) Created(ctx context.Context, author *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Do any local low-volatile operation before any TCP/UDP connection
	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorCreated)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Send domain event, spread aggregation side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, authorJSON, false)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"service":    e.ServiceName,
			"event_type": e.EventType,
			"priority":   e.Priority,
			"provider":   e.Provider,
		},
		BeforeSend: nil,
	}

	return topic.Send(ctx, m)
}

func (b *AuthorKafkaEventBus) StartUpdate(ctx context.Context, author *domain.Author, backup *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return err
	}

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorUpdatePending)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, authorJSON, false)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": uuid.New().String(),
			"operation":      domain.AuthorUpdated,
			"backup":         string(backupJSON),
			"service":        e.ServiceName,
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       e.Provider,
		},
		BeforeSend: nil,
	}

	return topic.Send(ctx, m)
}

func (b *AuthorKafkaEventBus) Updated(ctx context.Context, author *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorUpdated)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, authorJSON, true)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": uuid.New().String(),
			"service":        e.ServiceName,
			"type":           e.EventType,
			"priority":       e.Priority,
		},
		BeforeSend: nil,
	}

	return topic.Send(ctx, m)
}

func (b *AuthorKafkaEventBus) Deleted(ctx context.Context, id string, isHard bool) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	var topicName string
	if isHard {
		topicName = domain.AuthorHardDeleted
	} else {
		topicName = domain.AuthorDeleted
	}

	topic, err := eventbus.NewKafkaProducer(ctx, topicName)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Send domain event, Spread side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, []byte(id), false)
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"service":    e.ServiceName,
			"event_type": e.EventType,
			"priority":   e.Priority,
			"provider":   e.Provider,
		},
		BeforeSend: nil,
	})
}

func (b *AuthorKafkaEventBus) Restored(ctx context.Context, id string) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorRestored)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Send domain event, Spread side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, []byte(id), false)
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"service":    e.ServiceName,
			"event_type": e.EventType,
			"priority":   e.Priority,
			"provider":   e.Provider,
		},
		BeforeSend: nil,
	})
}

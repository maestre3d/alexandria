package infrastructure

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"gocloud.dev/pubsub"
	"strconv"
	"sync"
)

type AuthorKafkaEventBus struct {
	cfg *config.Kernel
	mtx *sync.Mutex
}

func NewAuthorKafkaEventBus(cfg *config.Kernel) *AuthorKafkaEventBus {
	return &AuthorKafkaEventBus{cfg, new(sync.Mutex)}
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

	var e *eventbus.Event
	var m *pubsub.Message
	if len(author.Owners) > 0 {
		// Send integration event, verify owner_id from identity usecase
		e = eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, authorJSON, true)
		m = &pubsub.Message{
			Body: e.Content,
			Metadata: map[string]string{
				"transaction_id": strconv.FormatUint(e.TransactionID, 10),
				"type":           e.EventType,
				"priority":       e.Priority,
			},
			BeforeSend: nil,
		}
	} else {
		// Send domain event, spread aggregation side-effects to all required services
		e = eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, authorJSON, false)
		m = &pubsub.Message{
			Body: e.Content,
			Metadata: map[string]string{
				"type":     e.EventType,
				"priority": e.Priority,
			},
			BeforeSend: nil,
		}
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
			"type":     e.EventType,
			"priority": e.EventType,
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
			"type":     e.EventType,
			"priority": e.EventType,
		},
		BeforeSend: nil,
	})
}

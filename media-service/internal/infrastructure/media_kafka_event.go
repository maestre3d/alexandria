package infrastructure

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"gocloud.dev/pubsub"
	"sync"
)

type MediaKafkaEvent struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewMediaKafakaEvent(cfg *config.Kernel) *MediaKafkaEvent {
	return &MediaKafkaEvent{mu: new(sync.Mutex), cfg: cfg}
}

func (e *MediaKafkaEvent) StartCreate(ctx context.Context, media domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ownerPool := make([]string, 0)
	ownerPool = append(ownerPool, media.PublisherID)

	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, ownerJSON)
	t := eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    media.ExternalID,
		SpanID:    "",
		TraceID:   "",
		Operation: domain.MediaCreated,
	}
	p, err := eventbus.NewKafkaProducer(ctx, domain.OwnerVerify)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"root_id":        t.RootID,
			"operation":      t.Operation,
			"service":        event.ServiceName,
			"event_type":     event.EventType,
			"priority":       event.Priority,
			"provider":       event.Provider,
		},
		BeforeSend: nil,
	}

	return p.Send(ctx, m)
}

func (e *MediaKafkaEvent) Created(ctx context.Context, media domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, mediaJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaCreated)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"service":    event.ServiceName,
			"event_type": event.EventType,
			"priority":   event.Priority,
			"provider":   event.Provider,
		},
		BeforeSend: nil,
	}

	return p.Send(ctx, m)
}

func (e *MediaKafkaEvent) StartUpdate(ctx context.Context, media domain.Media, backup domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ownerPool := make([]string, 0)
	ownerPool = append(ownerPool, media.PublisherID)
	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return err
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return err
	}

	t := &eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    media.ExternalID,
		SpanID:    "",
		TraceID:   "",
		Operation: domain.MediaUpdated,
		Backup:    string(backupJSON),
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, ownerJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.OwnerVerify)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"root_id":        t.RootID,
			"operation":      t.Operation,
			"backup":         t.Backup,
			"service":        event.ServiceName,
			"event_type":     event.EventType,
			"priority":       event.Priority,
			"provider":       event.Provider,
		},
		BeforeSend: nil,
	}

	return p.Send(ctx, m)
}

func (e *MediaKafkaEvent) Updated(ctx context.Context, media domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, mediaJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaUpdated)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"service":    event.ServiceName,
			"event_type": event.EventType,
			"priority":   event.Priority,
			"provider":   event.Provider,
		},
		BeforeSend: nil,
	}

	return p.Send(ctx, m)
}

func (e *MediaKafkaEvent) Removed(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaRemoved)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"service":    event.ServiceName,
			"event_type": event.EventType,
			"priority":   event.Priority,
			"provider":   event.Provider,
		},
		BeforeSend: nil,
	}

	return p.Send(ctx, m)
}

func (e *MediaKafkaEvent) Restored(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaRestored)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"service":    event.ServiceName,
			"event_type": event.EventType,
			"priority":   event.Priority,
			"provider":   event.Provider,
		},
		BeforeSend: nil,
	}

	return p.Send(ctx, m)
}

func (e *MediaKafkaEvent) HardRemoved(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaHardRemoved)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"service":    event.ServiceName,
			"event_type": event.EventType,
			"priority":   event.Priority,
			"provider":   event.Provider,
		},
		BeforeSend: nil,
	}

	return p.Send(ctx, m)
}

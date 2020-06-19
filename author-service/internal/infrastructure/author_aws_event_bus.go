package infrastructure

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"
	"os"
	"sync"
)

type AuthorAWSEventBus struct {
	cfg *config.Kernel
	mtx *sync.Mutex
}

func NewAuthorAWSEventBus(cfg *config.Kernel) *AuthorAWSEventBus {
	return &AuthorAWSEventBus{cfg, new(sync.Mutex)}
}

func (b *AuthorAWSEventBus) StartCreate(ctx context.Context, author *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	ownerPool := make([]string, 0)

	ownerPool = append(ownerPool, author.OwnerID)
	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return err
	}

	topic, err := pubsub.OpenTopic(ctx, os.Getenv("AWS_SNS_OWNER_PENDING"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	t := eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    author.ExternalID,
		SpanID:    "",
		TraceID:   "",
		Operation: domain.AuthorCreated,
		Backup:    "",
		Code:      0,
		Message:   "",
	}
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, ownerJSON)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"operation":      t.Operation,
			"service":        e.ServiceName,
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       e.Provider,
		},
		BeforeSend: nil,
	}

	return topic.Send(ctx, m)
}

func (b *AuthorAWSEventBus) Created(ctx context.Context, author *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Do any local low-volatile operation before any TCP/UDP connection
	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	// "awssns:///AWS_SNS_ARN:ALEXANDRIA_AUTHOR_CREATED?region=us-east-1"
	topic, err := pubsub.OpenTopic(ctx, os.Getenv("AWS_SNS_AUTHOR_CREATED"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Send domain event, spread aggregation side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, authorJSON)
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

func (b *AuthorAWSEventBus) StartUpdate(ctx context.Context, author *domain.Author, backup *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	ownerPool := make([]string, 0)

	ownerPool = append(ownerPool, author.OwnerID)
	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return err
	}

	backupJSON, err := json.Marshal(backup)
	if err != nil {
		return err
	}

	topic, err := pubsub.OpenTopic(ctx, os.Getenv("AWS_SNS_AUTHOR_PENDING"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	t := eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    author.ExternalID,
		SpanID:    "",
		TraceID:   "",
		Operation: domain.AuthorUpdated,
		Backup:    string(backupJSON),
		Code:      0,
		Message:   "",
	}
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, ownerJSON)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"operation":      t.Operation,
			"backup":         t.Backup,
			"service":        e.ServiceName,
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       e.Provider,
		},
		BeforeSend: nil,
	}

	return topic.Send(ctx, m)
}

func (b *AuthorAWSEventBus) Updated(ctx context.Context, author *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	topic, err := pubsub.OpenTopic(ctx, os.Getenv("AWS_SNS_AUTHOR_UPDATED"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityHigh, eventbus.ProviderAWS, authorJSON)
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

func (b *AuthorAWSEventBus) Deleted(ctx context.Context, id string, isHard bool) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Env must be like
	// "awssns:///AWS_SNS_ARN:ALEXANDRIA_AUTHOR_DELETED?region=us-east-1"
	var topicName string
	if isHard {
		topicName = os.Getenv("AWS_SNS_AUTHOR_PERMANENTLY_DELETED")
	} else {
		topicName = os.Getenv("AWS_SNS_AUTHOR_DELETED")
	}

	topic, err := pubsub.OpenTopic(ctx, topicName)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Send domain event, Spread side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderAWS, []byte(id))
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

func (b *AuthorAWSEventBus) Restored(ctx context.Context, id string) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Env must be like
	// "awssns:///AWS_SNS_ARN:ALEXANDRIA_AUTHOR_RESTORED?region=us-east-1"
	topic, err := pubsub.OpenTopic(ctx, os.Getenv("AWS_SNS_AUTHOR_RESTORED"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Send domain event, Spread side-effects to all required services
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderAWS, []byte(id))
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

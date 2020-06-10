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

func (b *AuthorAWSEventBus) Created(ctx context.Context, author *domain.Author) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// "awssns:///AWS_SNS_ARN:ALEXANDRIA_AUTHOR_CREATED?region=us-east-1"
	topic, err := pubsub.OpenTopic(ctx, os.Getenv("AWS_SNS_AUTHOR_CREATED"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	if len(author.Owners) > 0 {
		// Send integration event, verify owner_id from identity usecase
		// Use transaction message struct
		msg := struct {
			AuthorID  string          `json:"author_id"`
			OwnerPool []*domain.Owner `json:"owner_pool"`
		}{}

		msgJSON, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, msgJSON, true)
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

		err = topic.Send(ctx, m)
		if err != nil {
			return err
		}
	}

	// Send domain event, spread aggregation side-effects to all required services
	// Expose every claim available
	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderAWS, authorJSON, false)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"service":  e.ServiceName,
			"type":     e.EventType,
			"priority": e.Priority,
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

	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, authorJSON, true)
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
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderAWS, []byte(id), false)
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"service":  e.ServiceName,
			"priority": e.Priority,
			"type":     e.EventType,
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
	e := eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderAWS, []byte(id), false)
	return topic.Send(ctx, &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"service":  e.ServiceName,
			"priority": e.Priority,
			"type":     e.EventType,
		},
		BeforeSend: nil,
	})
}

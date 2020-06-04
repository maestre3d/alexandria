package infrastructure

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"
	"os"
	"strconv"
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

	var e *eventbus.Event
	var m *pubsub.Message
	if len(author.Owners) > 0 {
		// Send integration event, verify owner_id from identity usecase
		e = eventbus.NewEvent(b.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, authorJSON, true)
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
		e = eventbus.NewEvent(b.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderAWS, authorJSON, false)
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

func (b *AuthorAWSEventBus) Deleted(ctx context.Context, id string, isHard bool) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Env must be like
	// "awssns:///AWS_SNS_ARN:ALEXANDRIA_AUTHOR_DELETED?region=us-east-1"
	var topicName string
	if isHard {
		topicName = os.Getenv("AWS_SNS_AUTHOR_HARD_DELETED")
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
			"priority": e.Priority,
			"type":     e.EventType,
		},
		BeforeSend: nil,
	})
}

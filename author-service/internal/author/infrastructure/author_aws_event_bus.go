package infrastructure

import (
	"context"
	"encoding/json"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"
	"os"
	"sync"
)

type AuthorAWSEventBus struct {
	ctx context.Context
	mtx *sync.Mutex
}

func NewAuthorAWSEventBus(ctx context.Context) *AuthorAWSEventBus {
	return &AuthorAWSEventBus{ctx, new(sync.Mutex)}
}

func (b *AuthorAWSEventBus) AuthorCreated(author *domain.AuthorEntity) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// "awssns:///AWS_SNS_ARN:ALEXANDRIA_AUTHOR_CREATED?region=us-east-1"
	topic, err := pubsub.OpenTopic(b.ctx, os.Getenv("AWS_SNS_AUTHOR_CREATED"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(b.ctx)

	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	return topic.Send(b.ctx, &pubsub.Message{
		Body: authorJSON,
		Metadata: map[string]string{
			"importance": "mid",
			"type":       "integration",
		},
		BeforeSend: nil,
	})
}

func (b *AuthorAWSEventBus) AuthorDeleted(id string) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Env must be like
	// "awssns:///AWS_SNS_ARN:ALEXANDRIA_AUTHOR_DELETED?region=us-east-1"
	topic, err := pubsub.OpenTopic(b.ctx, os.Getenv("AWS_SNS_AUTHOR_DELETED"))
	if err != nil {
		return err
	}
	defer topic.Shutdown(b.ctx)

	return topic.Send(b.ctx, &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"importance": "mid",
			"type":       "integration",
		},
		BeforeSend: nil,
	})
}

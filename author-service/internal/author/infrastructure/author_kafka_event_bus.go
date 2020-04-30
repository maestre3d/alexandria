package infrastructure

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/sony/sonyflake"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"
)

type AuthorKafkaEventBus struct {
	ctx context.Context
	mtx *sync.Mutex
}

func NewAuthorKafkaEventBus(ctx context.Context) *AuthorKafkaEventBus {
	return &AuthorKafkaEventBus{ctx, new(sync.Mutex)}
}

func (b *AuthorKafkaEventBus) AuthorCreated(author *domain.AuthorEntity) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	authorJSON, err := json.Marshal(author)
	if err != nil {
		return err
	}

	topic, err := pubsub.OpenTopic(b.ctx, "kafka://ALEXANDRIA_AUTHOR_CREATED")
	if err != nil {
		return err
	}
	defer topic.Shutdown(b.ctx)

	// Create a transaction id using flake
	flake := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime:      time.Now(),
		MachineID:      nil,
		CheckMachineID: nil,
	})

	transactionID, err := flake.NextID()
	if err != nil {
		return err
	}

	return topic.Send(b.ctx, &pubsub.Message{
		Body: authorJSON,
		Metadata: map[string]string{
			"transaction_id": strconv.FormatUint(transactionID, 10),
			"type":           "domain_event",
			"importance":     "low",
		},
		BeforeSend: nil,
	})
}

func (b *AuthorKafkaEventBus) AuthorDeleted(id string) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	topic, err := pubsub.OpenTopic(b.ctx, "kafka://ALEXANDRIA_AUTHOR_DELETED")
	if err != nil {
		return err
	}
	defer topic.Shutdown(b.ctx)

	return topic.Send(b.ctx, &pubsub.Message{
		Body: []byte(id),
		Metadata: map[string]string{
			"type":       "domain_event",
			"importance": "low",
		},
		BeforeSend: nil,
	})
}

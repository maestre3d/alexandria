package transport

import (
	"context"

	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"strings"
)

func StartConsumers(ctx context.Context, cfg *config.KernelConfiguration) error {
	authorCreated, err := eventbus.NewKafkaConsumer(ctx, strings.ToUpper(cfg.Service), "ALEXANDRIA_AUTHOR_CREATED")
	if err != nil {
		return err
	}
	eventbus.ListenSubscriber(ctx, authorCreated)

	authorDeleted, err := eventbus.NewKafkaConsumer(ctx, strings.ToUpper(cfg.Service), "ALEXANDRIA_AUTHOR_DELETED")
	if err != nil {
		return err
	}
	eventbus.ListenSubscriber(ctx, authorDeleted)

	return nil
}

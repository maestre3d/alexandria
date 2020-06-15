package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/maestre3d/alexandria/author-service/cmd/identity-event-mock/usecasemock"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/event"
	"github.com/oklog/run"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "gocloud.dev/pubsub/kafkapubsub"
)

const (
	ServiceName = "identity"
)

func main() {
	// Required config
	// Root context
	ctx, cancel := context.WithCancel(context.Background())
	_ = os.Setenv("KAFKA_BROKERS", fmt.Sprintf("%s:%d", "localhost", 9092))

	var g run.Group
	{
		eventSrv := event.NewServer(ctx, cancel, consumeAuthorPending(ctx), consumeAuthorPending2(ctx))
		g.Add(func() error {
			log.Print("starting pubsub service")
			return eventSrv.Serve()
		}, func(error) {
			eventSrv.Close()
		})
	}
	{
		// Set up signal bind
		var (
			cancelInterrupt = make(chan struct{})
			c               = make(chan os.Signal, 2)
		)
		defer close(c)

		g.Add(func() error {
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			cancel()
			close(cancelInterrupt)
		})
	}

	_ = g.Run()
}

func consumeAuthorPending(ctx context.Context) *event.Consumer {
	sub, err := eventbus.NewKafkaConsumer(ctx, ServiceName, domain.AuthorPending)
	if err != nil {
		panic(err)
	}

	return &event.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    onAuthorCreated,
	}
}

func consumeAuthorPending2(ctx context.Context) *event.Consumer {
	sub, err := eventbus.NewKafkaConsumer(ctx, "identity2", domain.AuthorPending)
	if err != nil {
		panic(err)
	}

	return &event.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    onAuthorCreated,
	}
}

/* Handlers / Binder */

func onAuthorCreated(r *event.Request) error {
	author := new(domain.Author)
	err := json.Unmarshal(r.Message.Body, author)
	if err != nil {
		return err
	}

	log.Printf("%v", author)

	return usecasemock.ValidateOwners(r.Context, author, r.Message.Metadata["transaction_id"],
		r.Message.Metadata["operation"])
}

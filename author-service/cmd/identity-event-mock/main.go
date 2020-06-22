package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/maestre3d/alexandria/author-service/cmd/identity-event-mock/usecasemock"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
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
		eventSrv := eventbus.NewServer(ctx, consumeAuthorPending(ctx))
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

func consumeAuthorPending(ctx context.Context) *eventbus.Consumer {
	sub, err := eventbus.NewKafkaConsumer(ctx, ServiceName, domain.AuthorPending)
	if err != nil {
		panic(err)
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    onAuthorCreated,
	}
}

/* Handlers / Binder */

func onAuthorCreated(r *eventbus.Request) {
	author := new(domain.Author)
	err := json.Unmarshal(r.Message.Body, author)
	if err != nil {
		if r.Message.Nackable() {
			r.Message.Nack()
		}
		return
	}

	log.Printf("%v", author)

	err = usecasemock.ValidateOwners(r.Context, author, r.Message.Metadata["transaction_id"],
		r.Message.Metadata["operation"])
	if err != nil {
		if r.Message.Nackable() {
			r.Message.Nack()
		}
		return
	}

	r.Message.Ack()
}

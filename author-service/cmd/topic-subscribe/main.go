package main

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/config"
	logZap "github.com/go-kit/kit/log/zap"
	"github.com/oklog/run"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Init dependencies
	ctx := context.Background()

	loggerZap, _ := zap.NewProduction()
	defer loggerZap.Sync()
	level := zapcore.Level(8)
	logger := logZap.NewZapSugarLogger(loggerZap, level)

	cfg, _ := config.NewKernel(ctx)
	logger.Log("method", "main.topic-example", "msg", "kafka brokers set to "+cfg.EventBusConfig.KafkaHost)
	logger.Log("method", "main.topic-example", "msg", "dependencies loaded")

	var g run.Group
	{
		// Listen to Author created
		g.Add(func() error {
			// Env must be like
			// "awssqs://AWS_SQS_URL?region=us-east-1"
			subscriptionCreated, err := pubsub.OpenSubscription(ctx, fmt.Sprintf(`kafka://%s?topic=ALEXANDRIA_AUTHOR_CREATED`, "AUTHOR"))
			if err != nil {
				return err
			}
			listenQueue(ctx, subscriptionCreated)
			return nil
		}, func(err error) {
			log.Print(err)
			return
		})
	}
	{
		// Listen to Author Deleted
		g.Add(func() error {
			// Env must be like
			// "awssqs://AWS_SQS_URL?region=us-east-1"
			subscriptionDeleted, err := pubsub.OpenSubscription(ctx, fmt.Sprintf(`kafka://%s?topic=ALEXANDRIA_AUTHOR_DELETED`, "AUTHOR"))
			if err != nil {
				return err
			}

			listenQueue(ctx, subscriptionDeleted)
			return nil
		}, func(err error) {
			log.Print(err)
			return
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
			close(cancelInterrupt)
		})
	}

	g.Run()
}

func listenQueue(ctx context.Context, subscription *pubsub.Subscription) {
	defer subscription.Shutdown(ctx)

	// Loop on received messages. We can use a channel as a semaphore to limit how
	// many goroutines we have active at a time as well as wait on the goroutines
	// to finish before exiting.
	const maxHandlers = 10
	sem := make(chan struct{}, maxHandlers)
recvLoop:
	for {
		msg, err := subscription.Receive(ctx)
		if err != nil {
			// Errors from Receive indicate that Receive will no longer succeed.
			log.Printf("Receiving message: %v", err)
			break
		}

		// Wait if there are too many active handle goroutines and acquire the
		// semaphore. If the context is canceled, stop waiting and start shutting
		// down.
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break recvLoop
		}

		// Handle the message in a new goroutine.
		go func() {
			defer func() { <-sem }() // Release the semaphore.
			defer msg.Ack()          // Messages must always be acknowledged with Ack.

			// Do work based on the message, for example:
			fmt.Printf("Got message: %q\n", msg.Body)
			fmt.Printf("Message metadata: %q\n", msg.Metadata)
		}()
	}

	// We're no longer receiving messages. Wait to finish handling any
	// unacknowledged messages by totally acquiring the semaphore.
	for n := 0; n < maxHandlers; n++ {
		sem <- struct{}{}
	}
}

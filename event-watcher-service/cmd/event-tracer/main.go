package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-kit/kit/log"
	logZap "github.com/go-kit/kit/log/zap"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/shared/infrastructure/config"
	"github.com/oklog/run"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/kafkapubsub"
)

func main() {
	// Init dependencies
	ctx := context.Background()

	logger := NewLogger()

	cfg := config.NewKernelConfig(ctx, logger)
	logger.Log("method", "main.event-tracer", "msg", "kafka brokers set to "+cfg.EventBusConfig.KafkaHost)
	logger.Log("method", "main.event-tracer", "msg", "dependencies loaded")

	var g run.Group
	{
		// Listen to Author created
		g.Add(func() error {
			// Env must be like
			// "awssqs://AWS_SQS_URL?region=us-east-1"
			subscriptionCreated, err := pubsub.OpenSubscription(ctx, fmt.Sprintf(`kafka://%s?topic=ALEXANDRIA_AUTHOR_CREATED`, strings.ToUpper(cfg.Service)))
			if err != nil {
				return err
			}
			listenQueue(ctx, logger, subscriptionCreated)
			return nil
		}, func(err error) {
			logger.Log("err", err.Error())
			panic(err)
		})
	}
	{
		// Listen to Author Deleted
		g.Add(func() error {
			// Env must be like
			// "awssqs://AWS_SQS_URL?region=us-east-1"
			subscriptionDeleted, err := pubsub.OpenSubscription(ctx, fmt.Sprintf(`kafka://%s?topic=ALEXANDRIA_AUTHOR_DELETED`, strings.ToUpper(cfg.Service)))
			if err != nil {
				return err
			}

			listenQueue(ctx, logger, subscriptionDeleted)
			return nil
		}, func(err error) {
			logger.Log("err", err.Error())
			panic(err)
		})

	}
	{
		// Set up signal handler
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

func NewLogger() log.Logger {
	loggerZap, _ := zap.NewProduction()
	defer loggerZap.Sync()
	level := zapcore.Level(8)
	return logZap.NewZapSugarLogger(loggerZap, level)
}

func listenQueue(ctx context.Context, logger log.Logger, subscription *pubsub.Subscription) {
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
			logger.Log("msg", err.Error())
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
			logger.Log("msg", string(msg.Body))
			logger.Log("msg", fmt.Sprintf("%v", msg.Metadata))
		}()
	}

	// We're no longer receiving messages. Wait to finish handling any
	// unacknowledged messages by totally acquiring the semaphore.
	for n := 0; n < maxHandlers; n++ {
		sem <- struct{}{}
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/maestre3d/alexandria/event-telemetry-service/internal/shared/infrastructure/dependency"
	"github.com/maestre3d/alexandria/event-telemetry-service/pkg/transport"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexandria-oss/core/config"
	"github.com/go-kit/kit/log"
	logZap "github.com/go-kit/kit/log/zap"
	"github.com/oklog/run"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	_ "gocloud.dev/pubsub/kafkapubsub"
)

func main() {
	// Init dependencies
	ctx := context.Background()

	logger := NewLogger()
	cfg, err := config.NewKernelConfiguration(ctx)
	if err != nil {
		panic(err)
	}

	eventUseCase, cleanup, err := dependency.InjectEventUseCase()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	logger.Log("method", "main.event-telemetry", "msg", "kafka brokers set to "+cfg.EventBusConfig.KafkaHost)
	logger.Log("method", "main.event-telemetry", "msg", "dependencies loaded")

	var g run.Group

	{
		// Listen to Author created
		g.Add(func() error {
			// Env must be like
			// "awssqs://AWS_SQS_URL?region=us-east-1"
			return transport.StartConsumers(ctx, cfg)
		}, func(err error) {
			logger.Log("err", err.Error())
			panic(err)
		})
	}
	{

		srv := transport.NewHTTPServer(eventUseCase, cfg)
		l, err := net.Listen("tcp", srv.Addr)
		if err != nil {
			logger.Log("err", err.Error())
		}

		g.Add(func() error {
			return http.Serve(l, srv.Handler)
		}, func(err error) {
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

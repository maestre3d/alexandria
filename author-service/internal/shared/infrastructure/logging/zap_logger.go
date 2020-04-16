package logging

import (
	"go.uber.org/zap"
	"time"
)

// ZapLogger Uber's Zap Logger
type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger() (*ZapLogger, func(), error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, nil
	}

	cleanup := func() {
		err = logger.Sync()
	}

	logger.Info("logger started",
		zap.String("resource", "core.kernel.infrastructure.logging"),
		zap.Duration("backoff", time.Second),
	)

	return &ZapLogger{logger}, cleanup, nil
}

// Print Output a message along a resource location
func (z *ZapLogger) Print(message, resource string) {
	z.logger.Info(message,
		zap.String("resource", resource),
		zap.Duration("backoff", time.Second),
	)
}

// Error Output an error along a resource location
func (z *ZapLogger) Error(message, resource string) {
	z.logger.Error(message,
		zap.String("resource", resource),
		zap.Duration("backoff", time.Second),
	)
}

// Fatal Output and panic with an error along a resource location
func (z *ZapLogger) Fatal(message, resource string) {
	z.logger.Fatal(message,
		zap.String("resource", resource),
		zap.Duration("backoff", time.Second),
	)
}
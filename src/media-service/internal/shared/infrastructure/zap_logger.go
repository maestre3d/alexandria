package infrastructure

import (
	"go.uber.org/zap"
	"time"
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger() (*Logger, func() error, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		return logger.Sync()
	}

	rootLogger := &Logger{logger}

	return rootLogger, cleanup, nil
}

func (l *Logger) Print(message, resource string) {
	l.logger.Info(message,
		zap.String("resource", resource),
		zap.Duration("backoff", time.Second),
	)
}

func (l *Logger) Error(message, resource string) {
	l.logger.Error(message,
		zap.String("resource", resource),
		zap.Duration("backoff", time.Second),
	)
}

func (l *Logger) Fatal(message, resource string) {
	l.logger.Fatal(message,
		zap.String("resource", resource),
		zap.Duration("backoff", time.Second),
	)
}

func (l *Logger) Close() func() {
	return func() {
		err := l.logger.Sync()
		if err != nil {
			panic(err)
		}
	}
}

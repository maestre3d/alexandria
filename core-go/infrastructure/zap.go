package infrastructure

import (
	"github.com/go-kit/kit/log"
	logZap "github.com/go-kit/kit/log/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger() log.Logger {
	loggerZap, _ := zap.NewProduction()
	defer loggerZap.Sync()
	level := zapcore.Level(8)

	return logZap.NewZapSugarLogger(loggerZap, level)
}

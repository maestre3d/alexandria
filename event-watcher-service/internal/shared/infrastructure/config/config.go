package config

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
)

type KernelConfig struct {
	TransportConfig transportCfg
	MetricConfig    metricCfg
	EventBusConfig  eventBusCfg
	Version         string
	Service         string
}

type transportCfg struct {
	HTTPHost string
	HTTPPort int
	RPCHost  string
	RPCPort  int
}

type metricCfg struct {
	ZipkinHost     string
	ZipkinEndpoint string
	ZipkinBridge   bool
}

type eventBusCfg struct {
	KafkaHost string
	KafkaPort int
}

func NewKernelConfig(ctx context.Context, logger log.Logger) *KernelConfig {
	kernelConfig := new(KernelConfig)

	// Init config
	viper.SetConfigName("alexandria-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("/etc/alexandria/")
	viper.AddConfigPath("$HOME/.alexandria")
	viper.AddConfigPath(".")

	// Set default

	// Transport - HTTP
	viper.SetDefault("alexandria.service.transport.http.host", "0.0.0.0")
	viper.SetDefault("alexandria.service.transport.http.port", 8080)

	// Transport - RPC
	viper.SetDefault("alexandria.service.transport.rpc.host", "0.0.0.0")
	viper.SetDefault("alexandria.service.transport.rpc.port", 31337)

	// Tracing/Instrumentation
	viper.SetDefault("alexandria.tracing.zipkin.host", "http://localhost:9411/api/v2/spans")
	viper.SetDefault("alexandria.tracing.zipkin.endpoint", "0.0.0.0:8080")
	viper.SetDefault("alexandria.tracing.zipkin.bridge", true)

	// Event Bus (Message Brokers, Queues, Notifications)
	viper.SetDefault("alexandria.event.kafka.cluster.leader.host", "0.0.0.0")
	viper.SetDefault("alexandria.event.kafka.cluster.leader.port", 9092)

	// Service info
	viper.SetDefault("alexandria.info.version", "1.0.0")
	viper.SetDefault("alexandria.info.service", "event-tracer")

	// Open config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			err = viper.SafeWriteConfig()
			if err != nil {
				logger.Log(
					"method", "core.kernel.infrastructure.config",
					"msg", "configuration writing failed",
				)
			}
		} else {
			// Config file was found but another error was produced
			logger.Log(
				"method", "core.kernel.infrastructure.config",
				"msg", "default-local configuration used",
			)
		}
	}

	// Set up services ports
	kernelConfig.TransportConfig.HTTPHost = viper.GetString("alexandria.service.transport.http.host")
	kernelConfig.TransportConfig.HTTPPort = viper.GetInt("alexandria.service.transport.http.port")

	kernelConfig.TransportConfig.RPCHost = viper.GetString("alexandria.service.transport.rpc.host")
	kernelConfig.TransportConfig.RPCPort = viper.GetInt("alexandria.service.transport.rpc.port")

	kernelConfig.MetricConfig.ZipkinHost = viper.GetString("alexandria.tracing.zipkin.host")
	kernelConfig.MetricConfig.ZipkinEndpoint = viper.GetString("alexandria.tracing.zipkin.endpoint")
	kernelConfig.MetricConfig.ZipkinBridge = viper.GetBool("alexandria.tracing.zipkin.bridge")

	kernelConfig.EventBusConfig.KafkaHost = viper.GetString("alexandria.event.kafka.cluster.leader.host")
	kernelConfig.EventBusConfig.KafkaPort = viper.GetInt("alexandria.event.kafka.cluster.leader.port")

	kernelConfig.Version = viper.GetString("alexandria.info.version")
	kernelConfig.Service = viper.GetString("alexandria.info.service")

	// Prefer AWS KMS/Hashicorp Vault/Key Parameter Store over local, replace default or local config

	// Start up env
	os.Setenv("KAFKA_BROKERS", fmt.Sprintf("%s:%d", kernelConfig.EventBusConfig.KafkaHost, kernelConfig.EventBusConfig.KafkaPort))

	logger.Log(
		"method", "core.kernel.infrastructure.config",
		"msg", "kernel configuration created",
	)
	return kernelConfig
}

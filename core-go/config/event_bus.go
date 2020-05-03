package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type eventBus struct {
	KafkaHost string
	KafkaPort int
}

func setEventBusDefaultConfig() {
	viper.SetDefault("alexandria.event.kafka.cluster.leader.host", "0.0.0.0")
	viper.SetDefault("alexandria.event.kafka.cluster.leader.port", 9092)
}

func newEventBusConfig() eventBus {
	cfg := eventBus{
		KafkaHost: viper.GetString("alexandria.event.kafka.cluster.leader.host"),
		KafkaPort: viper.GetInt("alexandria.event.kafka.cluster.leader.port"),
	}

	// Start up required kafka env
	os.Setenv("KAFKA_BROKERS", fmt.Sprintf("%s:%d", cfg.KafkaHost, cfg.KafkaPort))

	return cfg
}

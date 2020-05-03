package config

import (
	"context"

	"github.com/spf13/viper"
)

// KernelConfiguration Alexandria kernel configuration struct
// Generates required OS env variables
type KernelConfiguration struct {
	TransportConfig transport
	TracingConfig   tracing

	EventBusConfig eventBus

	DocstoreConfig docstore
	DBMSConfig     dbms
	InMemoryConfig inMemory

	Version string
	Service string
}

// NewKernelConfiguration Generate a global configuration from alexandria-config.yml file
func NewKernelConfiguration(ctx context.Context) (*KernelConfiguration, error) {
	// Context is required to use gocloud.dev functions

	kernelConfig := new(KernelConfiguration)

	// Init config
	viper.SetConfigName("alexandria-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("/etc/alexandria/")
	viper.AddConfigPath("$HOME/.alexandria")
	viper.AddConfigPath(".")

	// Set default
	setTransportDefaultConfig()
	setTracingDefaultConfig()
	setEventBusDefaultConfig()
	setDocstoreDefaultConfig()
	setDBMSDefaultConfig()
	setInMemoryDefaultConfig()

	// Service info
	viper.SetDefault("alexandria.info.version", "")
	viper.SetDefault("alexandria.info.service", "")

	// Open config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			viper.SafeWriteConfig()
		}

		// Config file was found but another error was produced, use default values
	}

	// Map kernel configuration
	kernelConfig.TransportConfig = newTransportConfig()
	kernelConfig.TracingConfig = newTracingConfig()
	kernelConfig.EventBusConfig = newEventBusConfig()
	kernelConfig.DocstoreConfig = newDocstoreConfig()
	kernelConfig.DBMSConfig = newDBMSConfig()
	kernelConfig.InMemoryConfig = newInMemoryConfig()

	kernelConfig.Version = viper.GetString("alexandria.info.version")
	kernelConfig.Service = viper.GetString("alexandria.info.service")

	// Prefer AWS KMS/Hashicorp Vault/Key Parameter Store over local, replace default or local config
	// TODO: Implement Hashicorp Vault or AWS KMS key/secret fetching

	return kernelConfig, nil
}

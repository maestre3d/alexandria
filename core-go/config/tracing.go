package config

import "github.com/spf13/viper"

type tracing struct {
	ZipkinHost     string
	ZipkinEndpoint string
	ZipkinBridge   bool
}

func setTracingDefaultConfig() {
	viper.SetDefault("alexandria.tracing.zipkin.host", "http://localhost:9411/api/v2/spans")
	viper.SetDefault("alexandria.tracing.zipkin.endpoint", "0.0.0.0:8080")
	viper.SetDefault("alexandria.tracing.zipkin.bridge", true)
}

func newTracingConfig() tracing {
	return tracing{
		ZipkinHost:     viper.GetString("alexandria.tracing.zipkin.host"),
		ZipkinEndpoint: viper.GetString("alexandria.tracing.zipkin.endpoint"),
		ZipkinBridge:   viper.GetBool("alexandria.tracing.zipkin.bridge"),
	}
}

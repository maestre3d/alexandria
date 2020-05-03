package config

import "github.com/spf13/viper"

type transport struct {
	HTTPHost string
	HTTPPort int
	RPCHost  string
	RPCPort  int
}

func setTransportDefaultConfig() {
	// HTTP
	viper.SetDefault("alexandria.service.transport.http.host", "0.0.0.0")
	viper.SetDefault("alexandria.service.transport.http.port", 8080)

	// RPC
	viper.SetDefault("alexandria.service.transport.rpc.host", "0.0.0.0")
	viper.SetDefault("alexandria.service.transport.rpc.port", 31337)
}

func newTransportConfig() transport {
	return transport{
		HTTPHost: viper.GetString("alexandria.service.transport.http.host"),
		HTTPPort: viper.GetInt("alexandria.service.transport.http.port"),
		RPCHost:  viper.GetString("alexandria.service.transport.rpc.host"),
		RPCPort:  viper.GetInt("alexandria.service.transport.rpc.port"),
	}
}

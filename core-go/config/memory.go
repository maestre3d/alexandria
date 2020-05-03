package config

import "github.com/spf13/viper"

type inMemory struct {
	Network  string
	Host     string
	Port     int
	Password string
	Database string
}

func setInMemoryDefaultConfig() {
	viper.SetDefault("alexandria.persistence.mem.network", "")
	viper.SetDefault("alexandria.persistence.mem.host", "0.0.0.0")
	viper.SetDefault("alexandria.persistence.mem.port", 6379)
	viper.SetDefault("alexandria.persistence.mem.password", "")
	viper.SetDefault("alexandria.persistence.mem.database", "0")
}

func newInMemoryConfig() inMemory {
	return inMemory{
		Host:     viper.GetString("alexandria.persistence.mem.host"),
		Port:     viper.GetInt("alexandria.persistence.mem.port"),
		Password: viper.GetString("alexandria.persistence.mem.password"),
		Network:  viper.GetString("alexandria.persistence.mem.network"),
		Database: viper.GetString("alexandria.persistence.mem.database"),
	}
}

package config

import "github.com/spf13/viper"

type dbms struct {
	URL      string
	Driver   string
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

func setDBMSDefaultConfig() {
	viper.SetDefault("alexandria.persistence.dbms.url", "postgres://postgres:root@localhost/alexandria-media?sslmode=disable")
	viper.SetDefault("alexandria.persistence.dbms.driver", "postgres")
	viper.SetDefault("alexandria.persistence.dbms.user", "postgres")
	viper.SetDefault("alexandria.persistence.dbms.password", "root")
	viper.SetDefault("alexandria.persistence.dbms.host", "0.0.0.0")
	viper.SetDefault("alexandria.persistence.dbms.port", 5432)
	viper.SetDefault("alexandria.persistence.dbms.database", "alexandria-media")
}

func newDBMSConfig() dbms {
	return dbms{
		URL:      viper.GetString("alexandria.persistence.dbms.url"),
		Driver:   viper.GetString("alexandria.persistence.dbms.driver"),
		User:     viper.GetString("alexandria.persistence.dbms.user"),
		Password: viper.GetString("alexandria.persistence.dbms.password"),
		Host:     viper.GetString("alexandria.persistence.dbms.host"),
		Port:     viper.GetInt("alexandria.persistence.dbms.port"),
		Database: viper.GetString("alexandria.persistence.dbms.database"),
	}
}

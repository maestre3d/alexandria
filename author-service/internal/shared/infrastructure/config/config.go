package config

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
	"gocloud.dev/runtimevar"
)

type KernelConfig struct {
	HTTPPort string
	RPCPort string
	MainDBMSURL string
	MainMemHost string
	MainMemPassword string
	Version string
	Service string
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
	viper.SetDefault("alexandria.persistence.dbms.url", "postgres://postgres:root@localhost/alexandria-author?sslmode=disable")
	viper.SetDefault("alexandria.persistence.mem.host", "localhost:6379")
	viper.SetDefault("alexandria.persistence.mem.password", "")
	viper.SetDefault("alexandria.service.transport.http.port", ":8080")
	viper.SetDefault("alexandria.service.transport.rpc.port", ":31337")
	viper.SetDefault("alexandria.info.version", "1.0.0")
	viper.SetDefault("alexandria.info.service", "author")

	// Open config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			err = viper.SafeWriteConfig()
			if err != nil {
				logger.Log(
					"method", "core.kernel.infrastructure.config",
					"msg","configuration writing failed",
				)
			}
		} else {
			// Config file was found but another error was produced
			logger.Log(
				"method", "core.kernel.infrastructure.config",
				"msg","default-local configuration used",
			)
		}
	}

	// Set up services ports
	kernelConfig.HTTPPort = viper.GetString("alexandria.service.transport.http.port")
	kernelConfig.RPCPort = viper.GetString("alexandria.service.transport.rpc.port")

	kernelConfig.Version = viper.GetString("alexandria.info.version")

	// Prefer AWS KMS/Key Parameter Store over local
	// Get main DBMS connection string
	dbmsConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-dbms?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.MainDBMSURL = viper.GetString("alexandria.persistence.dbms.url")
		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg","dbms local url used",
		)
	} else if dbmsConn != nil {
		defer dbmsConn.Close()
		remoteVar, err := dbmsConn.Latest(ctx)
		if err == nil {
			kernelConfig.MainDBMSURL = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.MainDBMSURL = viper.GetString("alexandria.persistence.dbms.url")
		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg","dbms local url used",
		)
	}

	// Get main in-memory host string
	memConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-mem-host?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.MainMemHost = viper.GetString("alexandria.persistence.mem.host")
		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg","in-memory local host used",
		)
	} else if memConn != nil {
		defer memConn.Close()
		remoteVar, err := memConn.Latest(ctx)
		if err == nil {
			kernelConfig.MainMemHost = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.MainMemHost = viper.GetString("alexandria.persistence.mem.host")
		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg","in-memory local host used",
		)
	}

	// Get main in-memory password string
	memPassConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-mem-password?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.MainMemPassword = viper.GetString("alexandria.persistence.mem.password")
		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg","in-memory local password used",
		)
	} else if memPassConn != nil {
		defer memPassConn.Close()
		remoteVar, err := memPassConn.Latest(ctx)
		if err == nil {
			kernelConfig.MainMemPassword = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.MainMemPassword = viper.GetString("alexandria.persistence.mem.password")
		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg","in-memory local password used",
		)
	}

	logger.Log(
		"method", "core.kernel.infrastructure.config",
		"msg","kernel configuration created",
	)
	return kernelConfig
}

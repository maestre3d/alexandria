package config

import (
	"context"
	"github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"
	"github.com/spf13/viper"
	"gocloud.dev/runtimevar"
)

type KernelConfig struct {
	HTTPPort        string
	RPCPort         string
	MainDBMSURL     string
	MainMemHost     string
	MainMemPassword string
	Version         string
}

func NewKernelConfig(ctx context.Context, logger util.ILogger) *KernelConfig {
	kernelConfig := new(KernelConfig)

	// Init config
	viper.SetConfigName("alexandria-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/alexandria/")
	viper.AddConfigPath("$HOME/.alexandria")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")

	// Set default
	viper.SetDefault("alexandria.persistence.dbms.url", "postgres://postgres:root@localhost/alexandria-media?sslmode=disable")
	viper.SetDefault("alexandria.persistence.mem.host", "localhost:6379")
	viper.SetDefault("alexandria.persistence.mem.password", "")
	viper.SetDefault("alexandria.service.delivery.http.port", ":8080")
	viper.SetDefault("alexandria.service.delivery.rpc.port", ":31337")
	viper.SetDefault("alexandria.info.version", "1.0.0")

	// Open config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			err = viper.SafeWriteConfig()
			if err != nil {
				logger.Print("configuration writing failed", "kernel.infrastructure.config")
			}
		} else {
			// Config file was found but another error was produced
			logger.Print("default-local configuration used", "kernel.infrastructure.config")
		}
	}

	// Set up services ports
	kernelConfig.HTTPPort = viper.GetString("alexandria.service.delivery.http.port")
	kernelConfig.RPCPort = viper.GetString("alexandria.service.delivery.rpc.port")

	kernelConfig.Version = viper.GetString("alexandria.info.version")

	// Prefer AWS KMS/Key Parameter Store over local
	// Get main DBMS connection string
	dbmsConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-dbms?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.MainDBMSURL = viper.GetString("alexandria.persistence.dbms.url")
		logger.Print("dbms local url used", "kernel.infrastructure.config")
	} else if dbmsConn != nil {
		defer dbmsConn.Close()
		remoteVar, err := dbmsConn.Latest(ctx)
		if err == nil {
			kernelConfig.MainDBMSURL = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.MainDBMSURL = viper.GetString("alexandria.persistence.dbms.url")
		logger.Print("dbms local url used", "kernel.infrastructure.config")
	}

	// Get main in-memory host string
	memConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-mem-host?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.MainMemHost = viper.GetString("alexandria.persistence.mem.host")
		logger.Print("in-memory local host used", "kernel.infrastructure.config")
	} else if memConn != nil {
		defer memConn.Close()
		remoteVar, err := memConn.Latest(ctx)
		if err == nil {
			kernelConfig.MainMemHost = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.MainMemHost = viper.GetString("alexandria.persistence.mem.host")
		logger.Print("in-memory local host used", "kernel.infrastructure.config")
	}

	// Get main in-memory password string
	memPassConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-mem-password?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.MainMemPassword = viper.GetString("alexandria.persistence.mem.password")
		logger.Print("in-memory local password used", "kernel.infrastructure.config")
	} else if memPassConn != nil {
		defer memPassConn.Close()
		remoteVar, err := memPassConn.Latest(ctx)
		if err == nil {
			kernelConfig.MainMemPassword = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.MainMemPassword = viper.GetString("alexandria.persistence.mem.password")
		logger.Print("in-memory local password used", "kernel.infrastructure.config")
	}

	logger.Print("kernel configuration created", "kernel.infrastructure.config")
	return kernelConfig
}

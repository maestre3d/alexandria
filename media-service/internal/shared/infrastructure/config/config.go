package config

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
	"gocloud.dev/runtimevar"
)

type KernelConfig struct {
	TransportConfig transportCfg
	MetricConfig    metricCfg
	DBMSConfig      dbmsCfg
	MemConfig       memCfg
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

type dbmsCfg struct {
	URL      string
	Driver   string
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

type memCfg struct {
	Network  string
	Host     string
	Port     int
	Password string
	Database string
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
	// Data source - DBMS
	viper.SetDefault("alexandria.persistence.dbms.url", "postgres://postgres:root@localhost/alexandria-media?sslmode=disable")
	viper.SetDefault("alexandria.persistence.dbms.driver", "postgres")
	viper.SetDefault("alexandria.persistence.dbms.user", "postgres")
	viper.SetDefault("alexandria.persistence.dbms.password", "root")
	viper.SetDefault("alexandria.persistence.dbms.host", "0.0.0.0")
	viper.SetDefault("alexandria.persistence.dbms.port", 5432)
	viper.SetDefault("alexandria.persistence.dbms.database", "alexandria-media")

	// Data source - In-memory
	viper.SetDefault("alexandria.persistence.mem.network", "")
	viper.SetDefault("alexandria.persistence.mem.host", "0.0.0.0")
	viper.SetDefault("alexandria.persistence.mem.port", 6379)
	viper.SetDefault("alexandria.persistence.mem.password", "")
	viper.SetDefault("alexandria.persistence.mem.database", "0")

	// Transport - HTTP
	viper.SetDefault("alexandria.service.transport.http.host", "0.0.0.0")
	viper.SetDefault("alexandria.service.transport.http.port", 8080)

	// Transport - RPC
	viper.SetDefault("alexandria.service.transport.rpc.host", "0.0.0.0")
	viper.SetDefault("alexandria.service.transport.rpc.port", 31337)

	// Metric/Instrumentation
	viper.SetDefault("alexandria.service.metric.zipkin.host", "http://localhost:9411/api/v2/spans")
	viper.SetDefault("alexandria.service.metric.zipkin.endpoint", "0.0.0.0:8080")
	viper.SetDefault("alexandria.service.metric.zipkin.bridge", true)

	// Service info
	viper.SetDefault("alexandria.info.version", "1.0.0")
	viper.SetDefault("alexandria.info.service", "media")

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

	kernelConfig.MetricConfig.ZipkinHost = viper.GetString("alexandria.service.metric.zipkin.host")
	kernelConfig.MetricConfig.ZipkinEndpoint = viper.GetString("alexandria.service.metric.zipkin.endpoint")
	kernelConfig.MetricConfig.ZipkinBridge = viper.GetBool("alexandria.service.metric.zipkin.bridge")

	kernelConfig.Version = viper.GetString("alexandria.info.version")
	kernelConfig.Service = viper.GetString("alexandria.info.service")

	// Prefer AWS KMS/Key Parameter Store over local
	// Get main DBMS connection string
	dbmsConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-dbms?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.DBMSConfig.URL = viper.GetString("alexandria.persistence.dbms.url")
		kernelConfig.DBMSConfig.URL = viper.GetString("alexandria.persistence.dbms.url")
		kernelConfig.DBMSConfig.Driver = viper.GetString("alexandria.persistence.dbms.driver")
		kernelConfig.DBMSConfig.User = viper.GetString("alexandria.persistence.dbms.user")
		kernelConfig.DBMSConfig.Password = viper.GetString("alexandria.persistence.dbms.password")
		kernelConfig.DBMSConfig.Host = viper.GetString("alexandria.persistence.dbms.host")
		kernelConfig.DBMSConfig.Port = viper.GetInt("alexandria.persistence.dbms.port")
		kernelConfig.DBMSConfig.Database = viper.GetString("alexandria.persistence.dbms.database")

		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg", "dbms local url used",
		)
	} else if dbmsConn != nil {
		defer dbmsConn.Close()
		remoteVar, err := dbmsConn.Latest(ctx)
		if err == nil {
			kernelConfig.DBMSConfig.URL = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.DBMSConfig.URL = viper.GetString("alexandria.persistence.dbms.url")
		kernelConfig.DBMSConfig.Driver = viper.GetString("alexandria.persistence.dbms.driver")
		kernelConfig.DBMSConfig.User = viper.GetString("alexandria.persistence.dbms.user")
		kernelConfig.DBMSConfig.Password = viper.GetString("alexandria.persistence.dbms.password")
		kernelConfig.DBMSConfig.Host = viper.GetString("alexandria.persistence.dbms.host")
		kernelConfig.DBMSConfig.Port = viper.GetInt("alexandria.persistence.dbms.port")
		kernelConfig.DBMSConfig.Database = viper.GetString("alexandria.persistence.dbms.database")

		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg", "dbms local url used",
		)
	}

	// Get main in-memory host string
	memConn, err := runtimevar.OpenVariable(ctx, "awsparamstore://alexandria-persistence-mem-host?region=us-east-1&decoder=string")
	if err != nil {
		kernelConfig.MemConfig.Host = viper.GetString("alexandria.persistence.mem.host")
		kernelConfig.MemConfig.Port = viper.GetInt("alexandria.persistence.mem.port")
		kernelConfig.MemConfig.Password = viper.GetString("alexandria.persistence.mem.password")
		kernelConfig.MemConfig.Network = viper.GetString("alexandria.persistence.mem.network")
		kernelConfig.MemConfig.Database = viper.GetString("alexandria.persistence.mem.database")

		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg", "in-memory local host used",
		)
	} else if memConn != nil {
		defer memConn.Close()
		remoteVar, err := memConn.Latest(ctx)
		if err == nil {
			kernelConfig.MemConfig.Host = remoteVar.Value.(string)
		}
	} else {
		kernelConfig.MemConfig.Host = viper.GetString("alexandria.persistence.mem.host")
		kernelConfig.MemConfig.Port = viper.GetInt("alexandria.persistence.mem.port")
		kernelConfig.MemConfig.Password = viper.GetString("alexandria.persistence.mem.password")
		kernelConfig.MemConfig.Network = viper.GetString("alexandria.persistence.mem.network")
		kernelConfig.MemConfig.Database = viper.GetString("alexandria.persistence.mem.database")

		logger.Log(
			"method", "core.kernel.infrastructure.config",
			"msg", "in-memory local host used",
		)
	}

	logger.Log(
		"method", "core.kernel.infrastructure.config",
		"msg", "kernel configuration created",
	)
	return kernelConfig
}

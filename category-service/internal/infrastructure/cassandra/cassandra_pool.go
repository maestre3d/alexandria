package cassandra

import (
	"github.com/alexandria-oss/core/config"
	"github.com/gocql/gocql"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("alexandria.persistence.cassandra.cluster", []string{"cassandra"})
	viper.SetDefault("alexandria.persistence.cassandra.keyspace", "alexa1")
	viper.SetDefault("alexandria.persistence.cassandra.username", "")
	viper.SetDefault("alexandria.persistence.cassandra.password", "")
}

func NewCassandraPool(cfg *config.Kernel) *gocql.ClusterConfig {
	// TODO: Update kernel config to handle Apache Cassandra
	cluster := gocql.NewCluster(getCassandraCluster()[0])
	cluster.Hosts = getCassandraCluster()
	cluster.Authenticator = getCassandraAuthCreds()
	cluster.Consistency = gocql.Quorum
	cluster.PageSize = 100
	cluster.NumConns = 2

	// Shard context
	cluster.Keyspace = viper.GetString("alexandria.persistence.cassandra.keyspace")

	return cluster
}

func getCassandraCluster() []string {
	return viper.GetStringSlice("alexandria.persistence.cassandra.cluster")
}

func getCassandraAuthCreds() gocql.PasswordAuthenticator {
	return gocql.PasswordAuthenticator{
		Username: viper.GetString("alexandria.persistence.cassandra.username"),
		Password: viper.GetString("alexandria.persistence.cassandra.password"),
	}
}

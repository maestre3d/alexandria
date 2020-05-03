package config

import "github.com/spf13/viper"

type docstore struct {
	Collection   string
	PartitionKey string
}

func setDocstoreDefaultConfig() {
	viper.SetDefault("alexandria.persistence.doc.collection", "ALEXANDRIA_EVENT_WATCHER")
	viper.SetDefault("alexandria.persistence.doc.partition_key", "watcher_id")
}

func newDocstoreConfig() docstore {
	return docstore{
		Collection:   viper.GetString("alexandria.persistence.doc.collection"),
		PartitionKey: viper.GetString("alexandria.persistence.doc.partition_key"),
	}
}

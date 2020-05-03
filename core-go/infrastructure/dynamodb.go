package infrastructure

import (
	"context"
	"fmt"

	"github.com/maestre3d/alexandria/core-go/config"
	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/awsdynamodb"
)

// NewDynamoDBCollectionPool Obtain an AWS DynamoDB collection connection pool
func NewDynamoDBCollectionPool(ctx context.Context, cfg *config.KernelConfiguration) (*docstore.Collection, func(), error) {
	db, err := docstore.OpenCollection(ctx, fmt.Sprintf("dynamodb://%s?partition_key=%s", cfg.DocstoreConfig.Collection,
		cfg.DocstoreConfig.PartitionKey))
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		err = db.Close()
	}

	return db, cleanup, nil
}

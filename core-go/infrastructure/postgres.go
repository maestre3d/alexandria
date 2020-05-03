package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/maestre3d/alexandria/core-go/config"
	"gocloud.dev/postgres"
)

// NewPostgresPool Obtain a PostgreSQL connection pool
func NewPostgresPool(ctx context.Context, cfg *config.KernelConfiguration) (*sql.DB, func(), error) {
	db, err := postgres.Open(ctx, cfg.DBMSConfig.URL)
	if err != nil {
		return nil, nil, err
	}
	db.SetMaxOpenConns(50)
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxIdleConns(10)

	cleanup := func() {
		err = db.Close()
	}

	return db, cleanup, nil
}

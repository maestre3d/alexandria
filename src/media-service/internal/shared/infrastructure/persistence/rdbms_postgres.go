package persistence

import (
	"context"
	"database/sql"

	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"gocloud.dev/postgres"
)

func NewPostgresPool(ctx context.Context, logger util.ILogger) (*sql.DB, func() error, error) {
	db, err := postgres.Open(ctx, "postgres://postgres:root@localhost/alexandria-media?sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	logger.Print("main database started", "kernel.infrastructure.persistence")

	db.SetMaxOpenConns(50)
	closePool := func() error {
		return db.Close()
	}

	return db, closePool, nil
}

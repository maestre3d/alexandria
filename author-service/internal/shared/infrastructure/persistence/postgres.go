package persistence

import (
	"context"
	"database/sql"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
	"gocloud.dev/postgres"
	"time"
)

func NewPostgresPool(ctx context.Context, logger util.ILogger) (*sql.DB, func(), error) {
	db, err := postgres.Open(ctx, "postgres://postgres:root@localhost:5432/alexandria-author?sslmode=disable")
	if err != nil {
		return nil, nil, err
	}
	db.SetMaxOpenConns(50)
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxIdleConns(10)

	cleanup := func() {
		err = db.Close()
	}

	logger.Print("main dbms store started", "core.kernel.infrastructure.persistence")

	return db, cleanup, nil
}

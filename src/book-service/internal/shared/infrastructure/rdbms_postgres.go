package infrastructure

import (
	"context"
	"database/sql"

	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"gocloud.dev/postgres"
)

func NewPostgresPool(ctx context.Context, logger util.ILogger) (*sql.DB, func(), error) {
	db, err := postgres.Open(ctx, "postgres://postgres:root@localhost/alexandria-book?sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	db.SetMaxOpenConns(50)
	closePool := func() {
		err = db.Close()
		if err != nil {
			panic(err)
		}
	}

	return db, closePool, nil
}

package infrastructure

import (
	"context"
	"database/sql"
	"github.com/maestre3d/alexandria/src/book-service/internal/shared/domain/util"
	"gocloud.dev/postgres"
)

func NewPostgresPool(ctx context.Context, logger util.ILogger) (*sql.DB, func(), error) {
	db, err := postgres.Open(ctx, "postgres://root:root@localhost/alexandria-book")
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

func NewPostgresConnection(db *sql.DB, ctx context.Context) (*sql.Conn, func(), error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}

	closeConn := func() {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}

	return conn, closeConn, nil
}

package persistence

import (
	"context"
	"database/sql"
	"github.com/maestre3d/alexandria/src/media-service/internal/shared/infrastructure/config"

	"github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"
	"gocloud.dev/postgres"
)

func NewPostgresPool(ctx context.Context, logger util.ILogger, cfg *config.KernelConfig) (*sql.DB, func(), error) {
	db, err := postgres.Open(ctx, cfg.MainDBMSURL)
	if err != nil {
		return nil, nil, err
	}

	logger.Print("main database started", "kernel.infrastructure.persistence")

	db.SetMaxOpenConns(50)
	closePool := func() {
		err = db.Close()
	}

	return db, closePool, nil
}

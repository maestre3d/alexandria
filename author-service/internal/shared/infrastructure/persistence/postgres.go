package persistence

import (
	"context"
	"database/sql"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/shared/infrastructure/config"
	"gocloud.dev/postgres"
	"time"
)

func NewPostgresPool(ctx context.Context, logger log.Logger, cfg *config.KernelConfig) (*sql.DB, func(), error) {
	defer func(begin time.Time) {
		logger.Log(
			"method", "core.kernel.infrastructure.persistence",
			"msg", "main dbms database started",
			"took", time.Since(begin),
		)
	}(time.Now())

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

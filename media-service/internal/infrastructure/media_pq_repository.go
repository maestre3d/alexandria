package infrastructure

import (
	"context"
	"database/sql"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/lib/pq"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"strings"
	"sync"
)

type MediaPQRepository struct {
	db     *sql.DB
	mem    *redis.Client
	logger log.Logger
	mu     *sync.Mutex
}

func NewMediaPQRepository(db *sql.DB, mem *redis.Client, logger log.Logger) *MediaPQRepository {
	return &MediaPQRepository{
		db:     db,
		mem:    mem,
		logger: logger,
		mu:     new(sync.Mutex),
	}
}

func (r *MediaPQRepository) Save(ctx context.Context, media domain.Media) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.save", "db_connection", r.db.Stats().OpenConnections)

	statement := `INSERT INTO alexa1.media(external_id, title, display_name, description, language_code, publisher_id, author_id, publish_date, media_type) 
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = conn.ExecContext(ctx, statement, media.ExternalID, media.Title, media.DisplayName, media.Description, media.LanguageCode, media.PublisherID,
		media.AuthorID, media.PublishDate, media.MediaType)
	if err != nil {
		if customErr, ok := err.(*pq.Error); ok {
			if customErr.Code == "23505" {
				return exception.EntityExists
			}
		}
	}

	return err
}

func (r *MediaPQRepository) SaveRaw(ctx context.Context, media domain.Media) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.save_raw", "db_connection", r.db.Stats().OpenConnections)

	statement := `INSERT INTO alexa1.media 
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	_, err = conn.ExecContext(ctx, statement, media.ID, media.ExternalID, media.Title, media.DisplayName, media.Description, media.LanguageCode, media.PublisherID,
		media.AuthorID, media.PublishDate, media.MediaType, media.CreateTime, media.UpdateTime, media.DeleteTime, media.Active, media.ContentURL, media.TotalViews,
		media.Status)
	if err != nil {
		if customErr, ok := err.(*pq.Error); ok {
			if customErr.Code == "23505" {
				return exception.EntityExists
			}
		}
	}

	return err
}

func (r *MediaPQRepository) FetchByID(ctx context.Context, id string, showDisabled bool) (*domain.Media, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mem != nil {
		ctxR, cancel := context.WithCancel(ctx)
		defer cancel()

		authorChan := make(chan *domain.Media)
		defer close(authorChan)

		go func() {
			if author, ok := Get(ctxR, r.mem, id, "media").(*domain.Media); ok {
				authorChan <- author
				return
			}

			authorChan <- nil
		}()

		select {
		case author := <-authorChan:
			if author != nil {
				return author, nil
			}
		}
	}

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.fetch_by_id", "db_connection", r.db.Stats().OpenConnections)

	statement := `SELECT * FROM alexa1.media WHERE external_id = $1`
	if !showDisabled {
		statement += ` AND active = TRUE`
	}

	media := new(domain.Media)
	err = conn.QueryRowContext(ctx, statement, id).Scan(&media.ID, &media.ExternalID, &media.Title, &media.DisplayName, &media.Description,
		&media.LanguageCode, &media.PublisherID, &media.AuthorID, &media.PublishDate, &media.MediaType, &media.CreateTime, &media.UpdateTime,
		&media.DeleteTime, &media.Active, &media.ContentURL, &media.TotalViews, &media.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, exception.EntityNotFound
		}

		return nil, err
	}

	if r.mem != nil {
		ctxR, cancel := context.WithCancel(ctx)
		defer cancel()
		go Store(ctxR, r.mem, media.ExternalID, "media", media)
	}

	return media, nil
}

func (r *MediaPQRepository) Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*domain.Media, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.fetch", "db_connection", r.db.Stats().OpenConnections)

	// Query building
	b := &MediaQuery{Statement: `SELECT * FROM alexa1.media WHERE `}
	for key, value := range filter {
		switch {
		case key == "query" && value != "":
			b.Like(value).And()
			continue
		case key == "lang" && value != "":
			b.Language(value).And()
			continue
		case key == "publisher" && value != "":
			b.Publisher(value).And()
			continue
		case key == "author" && value != "":
			b.Author(value).And()
			continue
		case key == "media_type" && value != "":
			b.MediaType(value).And()
			continue
		}
	}

	isActive := "TRUE"
	if strings.ToUpper(filter["show_disabled"]) == "TRUE" {
		isActive = "FALSE"
	}

	sort := ""
	if strings.ToUpper(filter["sort"]) == "DESC" {
		sort = "DESC"
	} else if strings.ToUpper(filter["sort"]) == "ASC" {
		sort = "ASC"
	}

	if filter["filter_by"] == "id" {
		if params.Token != "" {
			b.Filter("id", ">=", params.Token, isActive).And()
		}

		b.Active(isActive).And().Raw("status = '"+domain.StatusDone+"'").OrderBy("id", "asc", sort)
	} else if filter["filter_by"] == "timestamp" {
		if params.Token != "" {
			b.Filter("update_time", "<=", params.Token, isActive).And()
		}

		b.Active(isActive).And().Raw("status = '"+domain.StatusDone+"'").OrderBy("update_time", "desc", sort)
	} else {
		if params.Token != "" {
			b.Filter("total_views", "<=", params.Token, isActive).And()
		}

		b.Active(isActive).And().Raw("status = '"+domain.StatusDone+"'").OrderBy("total_views", "desc", sort)
	}
	b.Limit(params.Size)

	// Query exec
	rows, err := conn.QueryContext(ctx, b.Statement)
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer func() {
		err = rows.Close()
	}()

	medias := make([]*domain.Media, 0)
	for rows.Next() {
		media := new(domain.Media)
		err = rows.Scan(&media.ID, &media.ExternalID, &media.Title, &media.DisplayName, &media.Description,
			&media.LanguageCode, &media.PublisherID, &media.AuthorID, &media.PublishDate, &media.MediaType, &media.CreateTime, &media.UpdateTime,
			&media.DeleteTime, &media.Active, &media.ContentURL, &media.TotalViews, &media.Status)
		if err != nil {
			return nil, err
		}
		medias = append(medias, media)
	}

	if len(medias) == 0 {
		return nil, exception.EntitiesNotFound
	}

	return medias, nil
}

func (r *MediaPQRepository) Replace(ctx context.Context, media domain.Media) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.replace", "db_connection", r.db.Stats().OpenConnections)

	statement := `UPDATE alexa1.media SET title = $1, display_name = $2, description = $3, language_code = $4, publisher_id = $5, author_id = $6, 
					publish_date = $7, media_type = $8, update_time = $9, content_url = $10, total_views = $11, status = $12 WHERE 
					external_id = $13 AND active = TRUE`
	row, err := conn.ExecContext(ctx, statement, media.Title, media.DisplayName, media.Description, media.LanguageCode, media.PublisherID, media.AuthorID,
		media.PublishDate, media.MediaType, media.UpdateTime, media.ContentURL, media.TotalViews, media.Status, media.ExternalID)
	if err != nil {
		if customErr, ok := err.(*pq.Error); ok {
			if customErr.Code == "23505" {
				return exception.EntityExists
			}
		}

		return err
	} else if af, err := row.RowsAffected(); af == 0 || err != nil {
		return exception.EntityNotFound
	}

	// write-through cache pattern
	if r.mem != nil {
		ctxR, cancel := context.WithCancel(ctx)
		defer cancel()
		go Store(ctxR, r.mem, media.ExternalID, "media", media)
	}

	return nil
}

func (r *MediaPQRepository) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.remove", "db_connection", r.db.Stats().OpenConnections)

	statement := `UPDATE alexa1.media SET active = FALSE, delete_time = CURRENT_TIMESTAMP WHERE external_id = $1 AND active = TRUE`
	res, err := conn.ExecContext(ctx, statement, id)
	if err != nil {
		return err
	} else if af, err := res.RowsAffected(); af == 0 || err != nil {
		return exception.EntityNotFound
	}

	if r.mem != nil {
		ctxR, cancel := context.WithCancel(ctx)
		defer cancel()

		go Remove(ctxR, r.mem, id, "media")
	}

	return nil
}

func (r *MediaPQRepository) Restore(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.restore", "db_connection", r.db.Stats().OpenConnections)

	statement := `UPDATE alexa1.media SET active = TRUE, delete_time = NULL WHERE external_id = $1 AND active = FALSE`
	res, err := conn.ExecContext(ctx, statement, id)
	if err != nil {
		return err
	} else if af, err := res.RowsAffected(); af == 0 || err != nil {
		return exception.EntityNotFound
	}

	if r.mem != nil {
		ctxR, cancel := context.WithCancel(ctx)
		defer cancel()

		go Remove(ctxR, r.mem, id, "media")
	}

	return nil
}

func (r *MediaPQRepository) HardRemove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.hard_remove", "db_connection", r.db.Stats().OpenConnections)

	statement := `DELETE FROM alexa1.media WHERE external_id = $1`
	res, err := conn.ExecContext(ctx, statement, id)
	if err != nil {
		return err
	} else if af, err := res.RowsAffected(); af == 0 || err != nil {
		return exception.EntityNotFound
	}

	if r.mem != nil {
		ctxR, cancel := context.WithCancel(ctx)
		defer cancel()

		go Remove(ctxR, r.mem, id, "media")
	}

	return nil
}

func (r *MediaPQRepository) ChangeState(ctx context.Context, id, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "media.infrastructure.postgres.change_state", "db_connection", r.db.Stats().OpenConnections)

	statement := `UPDATE alexa1.media SET status = $1 WHERE external_id = $2 AND active = TRUE`
	res, err := conn.ExecContext(ctx, statement, state, id)
	if err != nil {
		return err
	} else if af, err := res.RowsAffected(); af == 0 || err != nil {
		return exception.EntityNotFound
	}

	if r.mem != nil {
		ctxR, cancel := context.WithCancel(ctx)
		defer cancel()

		go Remove(ctxR, r.mem, id, "media")
	}

	return nil
}

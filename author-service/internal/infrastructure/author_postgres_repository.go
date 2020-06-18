package infrastructure

import (
	"context"
	"database/sql"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/lib/pq"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
)

// AuthorPostgresRepository DBMS Author repository
type AuthorPostgresRepository struct {
	db     *sql.DB
	mem    *redis.Client
	logger log.Logger
	mtx    *sync.Mutex
}

const tableName = "author"

// NewAuthorPostgresRepository Create an author repository
func NewAuthorPostgresRepository(dbPool *sql.DB, memPool *redis.Client, logger log.Logger) *AuthorPostgresRepository {
	return &AuthorPostgresRepository{
		db:     dbPool,
		mem:    memPool,
		logger: logger,
		mtx:    new(sync.Mutex),
	}
}

func (r *AuthorPostgresRepository) Save(ctx context.Context, author *domain.Author) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.save", "db_connection", r.db.Stats().OpenConnections)

	statement := `CALL alexa1.create_author($1, $2, $3, $4, $5, $6, $7)`

	_, err = conn.ExecContext(ctx, statement, author.ExternalID, author.FirstName, author.LastName, author.DisplayName, author.OwnershipType,
		author.OwnerID, author.Country)
	if err != nil {
		if customErr, ok := err.(*pq.Error); ok {
			if customErr.Code == "23505" {
				return exception.EntityExists
			}
		}
	}

	return err
}

func (r *AuthorPostgresRepository) SaveRaw(ctx context.Context, author *domain.Author) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.save_raw", "db_connection", r.db.Stats().OpenConnections)

	statement := `INSERT INTO alexa1.author VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err = conn.ExecContext(ctx, statement, author.ID, author.ExternalID, author.FirstName, author.LastName, author.DisplayName, author.OwnerID,
		author.OwnershipType, author.CreateTime, author.UpdateTime, author.DeleteTime, author.Active, author.Verified, author.Picture, author.TotalViews,
		author.Country, author.Status)
	if err != nil {
		if customErr, ok := err.(*pq.Error); ok {
			if customErr.Code == "23505" {
				return exception.EntityExists
			}
		}
	}

	return err
}

func (r *AuthorPostgresRepository) FetchByID(ctx context.Context, id string, showDisabled bool) (*domain.Author, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Cache-aside pattern
	if r.mem != nil {
		authorChan := make(chan *domain.Author)
		defer close(authorChan)

		go func() {
			if author, ok := Get(ctx, r.mem, id, tableName).(*domain.Author); ok {
				authorChan <- author
			} else {
				authorChan <- nil
			}
			return
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
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.fetch_by_id", "db_connection", r.db.Stats().OpenConnections)

	statement := `SELECT * FROM alexa1.author WHERE external_id = $1`
	if !showDisabled {
		statement += ` AND active = TRUE`
	}

	author := new(domain.Author)
	err = conn.QueryRowContext(ctx, statement, id).Scan(&author.ID, &author.ExternalID, &author.FirstName,
		&author.LastName, &author.DisplayName, &author.OwnerID, &author.OwnershipType, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
		&author.Active, &author.Verified, &author.Picture, &author.TotalViews, &author.Country, &author.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	// Write-through
	if r.mem != nil {
		go Store(ctx, r.mem, author.ExternalID, tableName, author)
	}

	return author, nil
}

func (r *AuthorPostgresRepository) Fetch(ctx context.Context, params *core.PaginationParams, filterParams core.FilterParams) ([]*domain.Author, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.fetch", "db_connection", r.db.Stats().OpenConnections)

	if params == nil {
		params = core.NewPaginationParams("", "0")
	}

	b := AuthorBuilder{`SELECT * FROM alexa1.author WHERE `}
	// Criteria map filter -> Query Builder
	for filterType, value := range filterParams {
		// Avoid nil values and comparison computation
		if value == "" {
			continue
		}
		switch {
		case filterType == "query":
			b.Query(value).And()
			continue
		case filterType == "display_name":
			b.DisplayName(value).And()
			continue
		case filterType == "ownership_type":
			// Avoid non-enum type values
			value = strings.ToLower(value)
			if value == string(domain.PrivateOwner) || value == string(domain.CommunityOwner) {
				b.Ownership(value).And()
			}
			continue
		case filterType == "owner_id":
			b.Owner(value).And()
			continue
		case filterType == "country":
			b.Country(value).And()
			continue
		}
	}

	isActive := "TRUE"
	if strings.ToUpper(filterParams["show_disabled"]) == "TRUE" {
		isActive = "FALSE"
	}

	sort := ""
	if strings.ToUpper(filterParams["sort"]) == "DESC" {
		sort = "DESC"
	} else if strings.ToUpper(filterParams["sort"]) == "ASC" {
		sort = "ASC"
	}

	// Keyset pagination, filtering type binding
	if filterParams["filter_by"] == "id" {
		// Filtering by ID
		if params.Token != "" {
			b.Filter("id", ">=", params.Token, isActive).And()
		}

		b.Active(isActive).And().Raw(`status = 'STATUS_DONE'`).OrderBy("id", "asc", sort)
	} else if filterParams["filter_by"] == "timestamp" {
		// Filtering by Timestamp
		if params.Token != "" {
			b.Filter("update_time", "<=", params.Token, isActive).And()
		}

		b.Active(isActive).And().Raw(`status = 'STATUS_DONE'`).OrderBy("update_time", "desc", sort)
	} else {
		// Filtering by Popularity - Default
		if params.Token != "" {
			b.Filter("total_views", "<=", params.Token, isActive).And()
		}

		b.Active(isActive).And().Raw(`status = 'STATUS_DONE'`).OrderBy("total_views", "desc", sort)
	}
	b.Limit(params.Size)

	// Query - entity mapping
	rows, err := conn.QueryContext(ctx, b.Statement)
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer func() {
		err = rows.Close()
	}()

	authors := make([]*domain.Author, 0)
	for rows.Next() {
		author := new(domain.Author)
		err = rows.Scan(&author.ID, &author.ExternalID, &author.FirstName,
			&author.LastName, &author.DisplayName, &author.OwnerID, &author.OwnershipType, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
			&author.Active, &author.Verified, &author.Picture, &author.TotalViews, &author.Country, &author.Status)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}

func (r *AuthorPostgresRepository) Replace(ctx context.Context, author *domain.Author) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.replace", "db_connection", r.db.Stats().OpenConnections)

	statement := `UPDATE alexa1.author SET first_name = $1, last_name = $2, display_name = $3, ownership_type = $4,
    update_time = $5, total_views = $6, owner_id = $7, status = $8, country = $9 WHERE external_id = $10 AND active = true`

	res, err := conn.ExecContext(ctx, statement, author.FirstName, author.LastName, author.DisplayName, author.OwnershipType, author.UpdateTime, author.TotalViews,
		author.OwnerID, author.Status, author.Country, author.ExternalID)
	if err != nil {
		if customErr, ok := err.(*pq.Error); ok {
			if customErr.Code == "23505" {
				return exception.EntityExists
			}
		}
		return err
	} else if affect, err := res.RowsAffected(); affect == 0 || err != nil {
		return exception.EntityNotFound
	}

	// write-through cache pattern
	if r.mem != nil {
		go Store(ctx, r.mem, author.ExternalID, tableName, author)
	}

	return nil
}

func (r *AuthorPostgresRepository) Remove(ctx context.Context, id string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.remove", "db_connection", r.db.Stats().OpenConnections)

	// Soft-delete
	statement := `UPDATE alexa1.author SET active = FALSE, delete_time = CURRENT_TIMESTAMP WHERE external_id = $1 AND active = TRUE`
	res, err := conn.ExecContext(ctx, statement, id)
	if err != nil {
		return err
	} else if affect, err := res.RowsAffected(); affect == 0 || err != nil {
		return exception.EntityNotFound
	}

	// write-through cache pattern
	if r.mem != nil {
		go Remove(ctx, r.mem, id, tableName)
	}

	return nil
}

func (r *AuthorPostgresRepository) Restore(ctx context.Context, id string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.restore", "db_connection", r.db.Stats().OpenConnections)

	statement := `UPDATE alexa1.author SET active = TRUE, delete_time = null WHERE external_id = $1 AND active = FALSE`
	res, err := conn.ExecContext(ctx, statement, id)
	if err != nil {
		return err
	} else if affect, err := res.RowsAffected(); affect == 0 || err != nil {
		return exception.EntityNotFound
	}

	// Invalidate cache
	// write-through cache pattern
	if r.mem != nil {
		go Remove(ctx, r.mem, id, tableName)
	}

	return nil
}

func (r *AuthorPostgresRepository) HardRemove(ctx context.Context, id string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.hard_remove", "db_connection", r.db.Stats().OpenConnections)

	// Hard-delete
	statement := `DELETE FROM alexa1.author WHERE external_id = $1`
	res, err := conn.ExecContext(ctx, statement, id)
	if err != nil {
		return err
	} else if affect, err := res.RowsAffected(); affect == 0 || err != nil {
		return exception.EntityNotFound
	}

	// write-through cache pattern
	if r.mem != nil {
		go Remove(ctx, r.mem, id, tableName)
	}

	return nil
}

func (r *AuthorPostgresRepository) ChangeState(ctx context.Context, id, state string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	_ = r.logger.Log("method", "author.infrastructure.postgres.change_state", "db_connection", r.db.Stats().OpenConnections)

	statement := `UPDATE alexa1.author SET status = $1 WHERE external_id = $2 AND active = TRUE`
	res, err := conn.ExecContext(ctx, statement, state, id)
	if err != nil {
		return err
	} else if affect, err := res.RowsAffected(); affect == 0 || err != nil {
		return exception.EntityNotFound
	}

	// write-through cache pattern
	if r.mem != nil {
		go Remove(ctx, r.mem, id, tableName)
	}

	return nil
}

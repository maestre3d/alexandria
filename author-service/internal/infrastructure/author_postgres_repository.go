package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"sync"
	"time"

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

	statement := `INSERT INTO AUTHOR(EXTERNAL_ID, FIRST_NAME, LAST_NAME, DISPLAY_NAME, BIRTH_DATE, OWNER_ID) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = conn.ExecContext(ctx, statement, author.ExternalID, author.FirstName, author.LastName, author.DisplayName, author.BirthDate, author.OwnerID)
	if customErr, ok := err.(*pq.Error); ok {
		if customErr.Code == "23505" {
			return exception.EntityExists
		}
	}

	return err
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

	author.UpdateTime = time.Now()

	statement := `UPDATE AUTHOR SET FIRST_NAME = $1, LAST_NAME = $2, DISPLAY_NAME = $3, BIRTH_DATE = $4, OWNER_ID = $5, STATUS = $6
	UPDATE_TIME = $7 WHERE EXTERNAL_ID = $8 AND DELETED = FALSE`

	_, err = conn.ExecContext(ctx, statement, author.FirstName, author.LastName, author.DisplayName, author.BirthDate,
		author.OwnerID, author.Status, author.UpdateTime, author.ExternalID)

	if customErr, ok := err.(*pq.Error); ok {
		if customErr.Code == "23505" {
			return exception.EntityExists
		}
	}
	if err != nil {
		return err
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
	statement := `UPDATE AUTHOR SET DELETED = TRUE WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`
	_, err = conn.ExecContext(ctx, statement, id)

	// write-through cache pattern
	if r.mem != nil {
		go Remove(ctx, r.mem, id, tableName)
	}

	return err
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

	statement := `UPDATE AUTHOR SET DELETED = FALSE WHERE EXTERNAL_ID = $1 AND DELETED = TRUE`
	_, err = conn.ExecContext(ctx, statement, id)

	return err
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
	statement := `DELETE FROM USERS WHERE EXTERNAL_ID = $1`
	_, err = conn.ExecContext(ctx, statement, id)

	// write-through cache pattern
	if r.mem != nil {
		go Remove(ctx, r.mem, id, tableName)
	}

	return err
}

func (r *AuthorPostgresRepository) FetchByID(ctx context.Context, id string) (*domain.Author, error) {
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
	_ = r.logger.Log("method", "author.infrastructure.postgres.fetchbyid", "db_connection", r.db.Stats().OpenConnections)

	// Add cache-aside pattern
	if r.mem != nil {
		authorChan := make(chan *domain.Author)

		go func() {
			if author, ok := Get(ctx, r.mem, id, tableName).(*domain.Author); ok {
				authorChan <- author
			}
		}()

		select {
		case author := <-authorChan:
			if author != nil {
				return author, nil
			}
		}
	}

	statement := `SELECT * FROM AUTHOR WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`

	author := new(domain.Author)
	err = conn.QueryRowContext(ctx, statement, id).Scan(&author.AuthorID, &author.ExternalID, &author.FirstName,
		&author.LastName, &author.DisplayName, &author.BirthDate, &author.OwnerID, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
		&author.Status, &author.Metadata, &author.Deleted)
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

	statement := `SELECT * FROM AUTHOR WHERE `

	// Criteria map filter -> Query Builder
	for filterType, value := range filterParams {
		switch {
		case filterType == "query" && value != "":
			statement += AndCriteriaSQL(QueryCriteriaSQL(value))
			continue
		case filterType == "display_name" && value != "":
			statement += AndCriteriaSQL(DisplayNameCriteriaSQL(value))
			continue
		case filterType == "owner_id" && value != "":
			statement += AndCriteriaSQL(OwnerIDCriteriaSQL(value))
		}
	}

	// Keyset
	if params.Token != "" {
		if filterParams["timestamp"] == "false" {
			statement += AndCriteriaSQL(fmt.Sprintf(`ID >= (SELECT ID FROM AUTHOR WHERE EXTERNAL_ID = '%s' AND DELETED = FALSE)`,
				params.Token))
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY ID ASC FETCH FIRST %d ROWS ONLY`, params.Size)
		} else if filterParams["timestamp"] == "" || filterParams["timestamp"] == "true" {
			// Timestamp/Most recent by default
			statement += AndCriteriaSQL(fmt.Sprintf(`UPDATE_TIME <= (SELECT UPDATE_TIME FROM AUTHOR WHERE EXTERNAL_ID = '%s' AND DELETED = FALSE)`,
				params.Token))
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY UPDATE_TIME DESC FETCH FIRST %d ROWS ONLY`, params.Size)
		}
	} else {
		if filterParams["timestamp"] == "false" {
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY ID ASC FETCH FIRST %d ROWS ONLY`, params.Size)
		} else if filterParams["timestamp"] == "" || filterParams["timestamp"] == "true" {
			// Timestamp/Most recent by default
			statement += fmt.Sprintf(`DELETED = FALSE ORDER BY UPDATE_TIME DESC FETCH FIRST %d ROWS ONLY`, params.Size)
		}
	}

	// Query - entity mapping
	rows, err := conn.QueryContext(ctx, statement)
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
		err = rows.Scan(&author.AuthorID, &author.ExternalID, &author.FirstName,
			&author.LastName, &author.DisplayName, &author.BirthDate, &author.OwnerID, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
			&author.Status, &author.Metadata, &author.Deleted)
		if err != nil {
			continue
		}
		authors = append(authors, author)
	}

	return authors, nil
}

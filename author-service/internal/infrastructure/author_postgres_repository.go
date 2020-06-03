package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
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

	var owner *domain.Owner
	for _, o := range author.Owners {
		if o.Role == string(domain.OwnerRole) {
			owner = o
		}
	}

	statement := `CALL alexa1.create_author($1, $2, $3, $4, $5, $6, $7)`

	if owner != nil {
		_, err = conn.ExecContext(ctx, statement, author.ExternalID, author.FirstName, author.LastName, author.DisplayName, author.OwnershipType,
			author.Owners[0].ID, author.Owners[0].Role)
		if err != nil {
			if customErr, ok := err.(*pq.Error); ok {
				if customErr.Code == "23505" {
					return exception.EntityExists
				}
			}
		}
	} else {
		return exception.NewErrorDescription(exception.RequiredField,
			fmt.Sprintf(exception.RequiredFieldString, "owner"))
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

	statement := `SELECT * FROM alexa1.author WHERE external_id = $1 AND active = TRUE`

	author := new(domain.Author)
	err = conn.QueryRowContext(ctx, statement, id).Scan(&author.ID, &author.ExternalID, &author.FirstName,
		&author.LastName, &author.DisplayName, &author.OwnershipType, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
		&author.Active, &author.Verified, &author.Picture)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	// Populate owners pool
	statement = `SELECT "user", role_type FROM alexa1.author_user WHERE fk_author = $1`
	rows, err := conn.QueryContext(ctx, statement)
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer func() {
		err = rows.Close()
	}()

	for rows.Next() {
		owner := new(domain.Owner)
		err = rows.Scan(&owner.ID, &owner.Role)
		if err != nil {
			return nil, err
		}
		author.Owners = append(author.Owners, owner)
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

	statement := `SELECT * FROM alexa1.author WHERE `

	// Criteria map filter -> Query Builder
	for filterType, value := range filterParams {
		switch {
		case filterType == "query" && value != "":
			statement += AndCriteriaSQL(QueryCriteriaSQL(value))
			continue
		case filterType == "display_name" && value != "":
			statement += AndCriteriaSQL(DisplayNameCriteriaSQL(value))
			continue
		case filterType == "ownership_type":
			statement += AndCriteriaSQL(OwnershipCriteriaSQL(value))
			continue
		case filterType == "owner_id":
			statement += AndCriteriaSQL(OwnerCriteriaSQL(value))
			continue
		}
	}

	// Keyset pagination
	if params.Token != "" {
		// If params contains any external_id token
		if filterParams["timestamp"] == "false" {
			statement += AndCriteriaSQL(fmt.Sprintf(`id >= (SELECT id FROM alexa1.author WHERE external_id = '%s' AND active = TRUE)`,
				params.Token))
			statement += fmt.Sprintf(`active = TRUE ORDER BY id ASC FETCH FIRST %d ROWS ONLY`, params.Size)
		} else {
			// Timestamp/Most recent by default
			statement += AndCriteriaSQL(fmt.Sprintf(`update_time <= (SELECT update_time FROM alexa1.author WHERE external_id = '%s' AND active = TRUE)`,
				params.Token))
			statement += fmt.Sprintf(`active = TRUE ORDER BY update_time DESC FETCH FIRST %d ROWS ONLY`, params.Size)
		}
	} else {
		if filterParams["timestamp"] == "false" {
			statement += fmt.Sprintf(`active = TRUE ORDER BY id ASC FETCH FIRST %d ROWS ONLY`, params.Size)
		} else {
			// Timestamp/Most recent by default
			statement += fmt.Sprintf(`active = TRUE ORDER BY update_time DESC FETCH FIRST %d ROWS ONLY`, params.Size)
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
		err = rows.Scan(&author.ID, &author.ExternalID, &author.FirstName,
			&author.LastName, &author.DisplayName, &author.OwnershipType, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
			&author.Active, &author.Verified, &author.Picture)
		if err != nil {
			continue
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

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer tx.Commit()

	statement := `CALL alexa1.update_author($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, statement, author.ExternalID, author.FirstName, author.LastName, author.DisplayName, author.OwnershipType)
	if err != nil {
		if customErr, ok := err.(*pq.Error); ok {
			if customErr.Code == "23505" {
				tx.Rollback()
				return exception.EntityExists
			}
		}

		tx.Rollback()
		return err
	}

	// Replace author's owners
	statement = `DELETE FROM alexa1.author_user WHERE external_id = $1`
	_, err = tx.ExecContext(ctx, statement, author.ExternalID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, owner := range author.Owners {
		statement = `CALL alexa1.add_user_author($1, $2, $3)`
		_, err := tx.ExecContext(ctx, statement, author.ExternalID, owner.ID, owner.Role)
		if err != nil {
			if customErr, ok := err.(*pq.Error); ok {
				if customErr.Code == "23505" {
					tx.Rollback()
					return exception.EntityExists
				}
			}
			tx.Rollback()
			return err
		}
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
	statement := `UPDATE alexa1.author SET active = FALSE WHERE external_id = $1 AND active = TRUE`
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

	statement := `UPDATE alexa1.author SET active = TRUE WHERE external_id = $1 AND active = FALSE`
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
	statement := `DELETE FROM alexa1.author WHERE external_id = $1`
	_, err = conn.ExecContext(ctx, statement, id)

	// write-through cache pattern
	if r.mem != nil {
		go Remove(ctx, r.mem, id, tableName)
	}

	return err
}

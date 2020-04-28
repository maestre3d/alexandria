package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/lib/pq"
	"github.com/maestre3d/alexandria/author-service/internal/author/domain"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/author-service/internal/shared/domain/util"
)

// AuthorDBMSRepository DBMS Author repository
type AuthorDBMSRepository struct {
	db     *sql.DB
	ctx    context.Context
	mem    *redis.Client
	logger log.Logger
	mtx    sync.RWMutex
}

// NewAuthorDBMSRepository Create an author repository
func NewAuthorDBMSRepository(dbPool *sql.DB, mem *redis.Client, ctx context.Context, logger log.Logger) *AuthorDBMSRepository {
	return &AuthorDBMSRepository{
		db:     dbPool,
		ctx:    ctx,
		mem:    mem,
		logger: logger,
	}
}

func (r *AuthorDBMSRepository) Save(author *domain.AuthorEntity) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	r.logger.Log("method", "author.infrastructure.dbmsrepository.save", "db_connection", r.db.Stats().OpenConnections)

	statement := `INSERT INTO AUTHOR(EXTERNAL_ID, FIRST_NAME, LAST_NAME, DISPLAY_NAME, BIRTH_DATE) VALUES ($1, $2, $3, $4, $5)`

	_, err = conn.ExecContext(r.ctx, statement, author.ExternalID, author.FirstName, author.LastName, author.DisplayName, author.BirthDate)
	if customErr, ok := err.(*pq.Error); ok {
		if customErr.Code == "23505" {
			return exception.EntityExists
		}
	}

	return err
}

func (r *AuthorDBMSRepository) Update(author *domain.AuthorEntity) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	r.logger.Log("method", "author.infrastructure.dbmsrepository.update", "db_connection", r.db.Stats().OpenConnections)

	author.UpdateTime = time.Now()

	statement := `UPDATE AUTHOR SET FIRST_NAME = $1, LAST_NAME = $2, DISPLAY_NAME = $3, BIRTH_DATE = $4,
	UPDATE_TIME = $5 WHERE EXTERNAL_ID = $6 AND DELETED = FALSE`

	_, err = conn.ExecContext(r.ctx, statement, author.FirstName, author.LastName, author.DisplayName, author.BirthDate,
		author.UpdateTime, author.ExternalID)

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
		go func() {
			memConn := r.mem.Conn()
			defer func() {
				err = memConn.Close()
			}()

			err := memConn.Get("author:" + author.ExternalID).Err()
			if err == nil {
				authorJSON, err := json.Marshal(author)
				if err == nil {
					err = memConn.Set("author:"+author.ExternalID, authorJSON, (time.Hour * 24)).Err()
				}
			}
		}()
	}

	return nil
}

func (r *AuthorDBMSRepository) Remove(id string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	r.logger.Log("method", "author.infrastructure.dbmsrepository.remove", "db_connection", r.db.Stats().OpenConnections)

	// Soft-delete
	statement := `UPDATE AUTHOR SET DELETED = TRUE WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`
	_, err = conn.ExecContext(r.ctx, statement, id)

	// write-through cache pattern
	if r.mem != nil {
		go func() {
			memConn := r.mem.Conn()
			defer func() {
				err = memConn.Close()
			}()

			err := memConn.Get(fmt.Sprintf("author:%s", id)).Err()
			if err == nil {
				err = memConn.Del("author:" + id).Err()
			}
		}()
	}

	return err
}

func (r *AuthorDBMSRepository) FetchByID(id string) (*domain.AuthorEntity, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	r.logger.Log("method", "author.infrastructure.dbmsrepository.fetchbyid", "db_connection", r.db.Stats().OpenConnections)

	// Add cache-aside pattern
	if r.mem != nil {
		authorChan := make(chan *domain.AuthorEntity)

		go func() {
			memCon := r.mem.Conn()
			defer func() {
				err := memCon.Close()
				if err != nil {
					authorChan <- nil
				}
			}()

			authorMem, err := memCon.Get(fmt.Sprintf("author:%s", id)).Result()
			if err == nil {
				author := new(domain.AuthorEntity)
				err = json.Unmarshal([]byte(authorMem), author)
				if err == nil {
					authorChan <- author
				}

				authorChan <- nil
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

	statement := `SELECT * FROM AUTHOR WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`

	author := new(domain.AuthorEntity)
	err = conn.QueryRowContext(r.ctx, statement, id).Scan(&author.AuthorID, &author.ExternalID, &author.FirstName,
		&author.LastName, &author.DisplayName, &author.BirthDate, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
		&author.Metadata, &author.Deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	// Write-through
	if r.mem != nil {
		go func() {
			memConn := r.mem.Conn()
			defer func() {
				err = memConn.Close()
			}()

			authorJSON, err := json.Marshal(author)
			if err == nil {
				err = memConn.Set(fmt.Sprintf("author:%s", id), authorJSON, (time.Hour * 24)).Err()
			}
		}()
	}

	return author, nil
}

func (r *AuthorDBMSRepository) Fetch(params *util.PaginationParams, filterParams util.FilterParams) ([]*domain.AuthorEntity, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = conn.Close()
	}()
	// Use Go CDK OpenCensus database metrics
	r.logger.Log("method", "author.infrastructure.dbmsrepository.fetch", "db_connection", r.db.Stats().OpenConnections)

	if params == nil {
		params = util.NewPaginationParams("", "0")
	}

	statement := `SELECT * FROM AUTHOR WHERE `
	params.Size += 1

	// Criteria map filter -> Query Builder
	for filterType, value := range filterParams {
		switch {
		case filterType == "query" && value != "":
			statement += AndCriteriaSQL(QueryCriteriaSQL(value))
			continue
		case filterType == "display_name" && value != "":
			statement += AndCriteriaSQL(DisplayNameCriteriaSQL(value))
			continue
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
	rows, err := conn.QueryContext(r.ctx, statement)
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer func() {
		err = rows.Close()
	}()

	authors := make([]*domain.AuthorEntity, 0)
	for rows.Next() {
		author := new(domain.AuthorEntity)
		err = rows.Scan(&author.AuthorID, &author.ExternalID, &author.FirstName,
			&author.LastName, &author.DisplayName, &author.BirthDate, &author.CreateTime, &author.UpdateTime, &author.DeleteTime,
			&author.Metadata, &author.Deleted)
		if err != nil {
			continue
		}
		authors = append(authors, author)
	}

	return authors, nil
}

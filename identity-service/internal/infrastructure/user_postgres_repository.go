package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/lib/pq"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"sync"
	"time"
)

type UserPostgresRepository struct {
	logger log.Logger
	db     *sql.DB
	mem    *redis.Client
	mtx    *sync.Mutex
}

func NewUserPostgresRepository(logger log.Logger, db *sql.DB) *UserPostgresRepository {
	return &UserPostgresRepository{
		logger: logger,
		db:     db,
		mem:    nil,
		mtx:    new(sync.Mutex),
	}
}

// SetInMem optional runtime in-memory persistence engine injection
func (r *UserPostgresRepository) SetInMem(mem *redis.Client) {
	r.mem = mem
}

func (r *UserPostgresRepository) Save(ctx context.Context, user domain.User) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	query := `INSERT INTO USERS(ID, EXTERNAL_ID, USERNAME, PASSWORD, NAME, LAST_NAME, EMAIL, GENDER,
				LOCALE, ROLE) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = conn.ExecContext(ctx, query, user.ID, user.ExternalID, user.Username, user.Password, user.Name, user.LastName,
		user.Email, user.Gender, user.Locale, user.Role)
	if customErr, ok := err.(*pq.Error); ok {
		if customErr.Code == "23505" {
			return exception.EntityExists
		}
	}

	return err
}

func (r *UserPostgresRepository) Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*domain.User, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	query := fmt.Sprintf(`SELECT * FROM USERS WHERE `)

	// Iterate through filter params
	for k, v := range filter {
		switch k {
		case "query":
			query += AndCriteriaSQL(UserQueryCriteria(v))
			continue
		}
	}

	if params.Token != "" {
		query += AndCriteriaSQL(fmt.Sprintf(`ID >= (SELECT ID FROM USERS WHERE EXTERNAL_ID = '%s' AND DELETED = FALSE AND ACTIVE = TRUE)`,
			params.Token))
	}

	if filter["timestamp"] == "false" {
		query += fmt.Sprintf(`DELETED = FALSE AND ACTIVE = TRUE ORDER BY ID ASC FETCH FIRST %d ROWS ONLY`, params.Size)
	} else {
		// Using timestamp
		query += fmt.Sprintf(`DELETED = FALSE AND ACTIVE = TRUE ORDER BY UPDATE_TIME DESC FETCH FIRST %d ROWS ONLY`, params.Size)
	}

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := new(domain.User)
		err = rows.Scan(&user.ID, &user.ExternalID, &user.Username, &user.Password, &user.Name, &user.LastName, &user.Email,
			&user.Gender, &user.Locale, &user.Picture, &user.Role, &user.CreateTime, &user.UpdateTime,
			&user.DeleteTime, &user.Deleted, &user.Active, &user.Verified)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserPostgresRepository) FetchOne(ctx context.Context, token string) (*domain.User, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.mem != nil {
		// TODO: Add cache-aside pattern
	}

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	row := conn.QueryRowContext(ctx, `SELECT * FROM USERS WHERE EXTERNAL_ID = $1 OR USERNAME = $1 AND DELETED = FALSE AND ACTIVE = TRUE`, token)

	user := new(domain.User)

	err = row.Scan(&user.ID, &user.ExternalID, &user.Username, &user.Password, &user.Name, &user.LastName, &user.Email,
		&user.Gender, &user.Locale, &user.Picture, &user.Role, &user.CreateTime, &user.UpdateTime,
		&user.DeleteTime, &user.Deleted, &user.Active, &user.Verified)
	if err != nil {
		return nil, err
	}

	if r.mem != nil {
		// TODO: Add write-through pattern
	}

	return user, nil
}

func (r *UserPostgresRepository) Replace(ctx context.Context, user domain.User) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.ExecContext(ctx, `UPDATE USERS SET USERNAME = $1, PASSWORD = $2, NAME = $2, LAST_NAME = $3, EMAIL = $4, GENDER = $5, 
		LOCALE = $6, UPDATE_TIME =$7, PICTURE = $8, ROLE = $9, ACTIVE = $10, VERIFIED = $11 WHERE EXTERNAL_ID = $12 AND DELETED = FALSE 
		AND ACTIVE = TRUE`,
		user.Username, user.Password, user.Name, user.LastName, user.Email, user.Gender, user.Locale,
		user.UpdateTime, user.Picture, user.Role, user.Active, user.Verified, user.ExternalID)
	if customErr, ok := err.(*pq.Error); ok {
		if customErr.Code == "23505" {
			return exception.EntityExists
		}
	} else if err != nil {
		return err
	}

	if r.mem != nil {
		// TODO: Add write-through pattern
	}

	return nil
}

func (r *UserPostgresRepository) Remove(ctx context.Context, token string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.ExecContext(ctx, `UPDATE USERS SET DELETED = TRUE, DELETE_TIME = $1 WHERE EXTERNAL_ID = $2 OR USERNAME = $2 AND DELETED = FALSE`,
		time.Now(), token)
	if err != nil {
		return err
	}

	if r.mem != nil {
		// TODO: Add write-through pattern
	}

	return nil
}

func (r *UserPostgresRepository) Restore(ctx context.Context, token string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.ExecContext(ctx, `UPDATE USERS SET DELETED = FALSE, DELETE_TIME = NULL WHERE EXTERNAL_ID = $1 OR USERNAME = $1 AND DELETED = TRUE`,
		token)
	return err
}

func (r *UserPostgresRepository) HardRemove(ctx context.Context, token string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.ExecContext(ctx, `DELETE FROM USERS WHERE EXTERNAL_ID = $1 OR USERNAME = $1`,
		time.Now(), token)
	if err != nil {
		return err
	}

	if r.mem != nil {
		// TODO: Add write-through pattern
	}

	return nil
}

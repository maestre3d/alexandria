package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"sync"
	"time"
)

type UserPostgresRepository struct {
	ctx    context.Context
	logger log.Logger
	mtx    *sync.Mutex
	db     *sql.DB
	mem    *redis.Client
}

func NewUserPostgresRepository(ctx context.Context, logger log.Logger, db *sql.DB, mem *redis.Client) *UserPostgresRepository {
	return &UserPostgresRepository{
		ctx:    ctx,
		logger: logger,
		mtx:    new(sync.Mutex),
		db:     db,
		mem:    mem,
	}
}

func (r *UserPostgresRepository) Save(user domain.User) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.mem != nil {
		// TODO: Add possible in-memory transaction(s)
	}

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `INSERT INTO USERS(EXTERNAL_ID, USERNAME, NAME, LAST_NAME, EMAIL, GENDER,
				LOCALE) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = conn.ExecContext(r.ctx, query, user.ExternalID, user.Username, user.Name, user.LastName,
		user.Email, user.Gender, user.Locale)
	return err
}

func (r *UserPostgresRepository) Fetch(params core.PaginationParams, filter core.FilterParams) ([]*domain.User, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := fmt.Sprintf(`SELECT * FROM USERS WHERE DELETED = FALSE`)

	rows, err := conn.QueryContext(r.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := new(domain.User)
		err = rows.Scan(&user.ID, &user.ExternalID, &user.Username, &user.Name, &user.LastName, &user.Email,
			&user.Gender, &user.Locale, &user.Picture, &user.Deleted, &user.CreatedAt, &user.UpdatedAt,
			&user.DeletedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserPostgresRepository) FetchByID(id string) (*domain.User, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.mem != nil {
		// TODO: Add cache-aside pattern
	}

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	row := conn.QueryRowContext(r.ctx, `SELECT * FROM USERS WHERE EXTERNAL_ID = $1 AND DELETED = FALSE`, id)

	user := new(domain.User)

	err = row.Scan(&user.ID, &user.ExternalID, &user.Username, &user.Name, &user.LastName, &user.Email,
		&user.Gender, &user.Locale, &user.Picture, &user.Deleted, &user.CreatedAt, &user.UpdatedAt,
		&user.DeletedAt)
	if err != nil {
		return nil, err
	}

	if r.mem != nil {
		// TODO: Add write-through pattern
	}

	return user, nil
}

func (r *UserPostgresRepository) Replace(user domain.User) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(r.ctx, `UPDATE USERS SET USERNAME = $1, NAME = $2, LAST_NAME = $3, EMAIL = $4, GENDER = $5,
				LOCALE = $6, UPDATED_AT =$7 WHERE EXTERNAL_ID = $8 AND DELETED = FALSE`, user.Username, user.Name, user.LastName,
		user.Email, user.Gender, user.Locale, user.UpdatedAt, user.ExternalID)
	if err != nil {
		return err
	}

	if r.mem != nil {
		// TODO: Add write-through pattern
	}

	return nil
}

func (r *UserPostgresRepository) Remove(id string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	conn, err := r.db.Conn(r.ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(r.ctx, `UPDATE USERS SET DELETED = TRUE, DELETED_AT = $1 WHERE EXTERNAL_ID = $2 AND DELETED = FALSE`,
		time.Now(), id)
	if err != nil {
		return err
	}

	if r.mem != nil {
		// TODO: Add write-through pattern
	}

	return nil
}

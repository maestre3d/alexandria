package infrastructure

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/gocql/gocql"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"sync"
)

type CategoryRepositoryCassandra struct {
	mu   *sync.RWMutex
	pool *gocql.ClusterConfig
}

func NewCategoryRepositoryCassandra(pool *gocql.ClusterConfig) *CategoryRepositoryCassandra {
	return &CategoryRepositoryCassandra{
		mu:   new(sync.RWMutex),
		pool: pool,
	}
}

func (r *CategoryRepositoryCassandra) Save(ctx context.Context, category domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	categoryExists := new(domain.Category)
	err = s.Query(`SELECT external_id FROM alexa1.category WHERE category_name = ? LIMIT 1 ALLOW FILTERING`, category.Name).Consistency(gocql.One).
		Scan(&categoryExists.ExternalID)
	if err != nil {
		if err != gocql.ErrNotFound {
			return err
		}
	} else if categoryExists.ExternalID != "" {
		return exception.EntityExists
	}

	err = s.Query(`INSERT INTO alexa1.category (id, external_id, category_name, create_time, update_time, active) VALUES  
		(?, ?, ?, ?, ?, ?)`, gocql.TimeUUID(), category.ExternalID, category.Name, category.CreateTime, category.UpdateTime, category.Active).
		WithContext(ctx).Exec()

	return err
}

func (r *CategoryRepositoryCassandra) FetchByID(ctx context.Context, id string, activeOnly bool) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	statement := `SELECT * FROM alexa1.category WHERE TOKEN(external_id) = TOKEN(?)`
	if activeOnly {
		statement += ` AND active = true `
	}
	statement += ` LIMIT 1 ALLOW FILTERING`
	category := new(domain.Category)
	err = s.Query(statement,
		id).Consistency(gocql.One).WithContext(ctx).
		Scan(&category.ExternalID, &category.ID, &category.Active, &category.Name, &category.CreateTime, &category.UpdateTime)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, exception.EntityNotFound
		}
		return nil, err
	}

	return category, nil
}

func (r *CategoryRepositoryCassandra) Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	builder := &CategoryCassandraBuilder{Statement: `SELECT * FROM alexa1.category WHERE `}
	for k, v := range filter {
		switch {
		case k == "query" && v != "":
			builder.Name(v).And()
			continue
		}
	}

	if params.Token != "" {
		builder.Statement += `TOKEN(external_id) >= TOKEN('` + params.Token + `')`
		builder.And()
	}

	if filter["show_disabled"] != "" {
		builder.Active(false)
	} else {
		builder.Active(true)
	}

	builder.Limit(params.Size)
	builder.Statement += ` ALLOW FILTERING`

	iter := s.Query(builder.Statement).WithContext(ctx).PageSize(params.Size).Iter()
	if iter.NumRows() == 0 {
		return nil, exception.EntitiesNotFound
	}

	category := domain.Category{}
	categories := make([]*domain.Category, 0)
	for iter.Scan(&category.ExternalID, &category.ID, &category.Active, &category.Name, &category.CreateTime, &category.UpdateTime) {
		catMemento := category
		categories = append(categories, &catMemento)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepositoryCassandra) Replace(ctx context.Context, category domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	categoryExists := new(domain.Category)
	err = s.Query(`SELECT external_id FROM alexa1.category WHERE category_name = ? LIMIT 1 ALLOW FILTERING`, category.Name).Consistency(gocql.One).
		Scan(&categoryExists.ExternalID)
	if err != nil {
		if err != gocql.ErrNotFound {
			return err
		}
	} else if categoryExists.ExternalID != "" {
		return exception.EntityExists
	}

	err = s.Query(`UPDATE alexa1.category SET category_name = ?, update_time = ? WHERE external_id = ? AND id = ?`, category.Name,
		category.UpdateTime, category.ExternalID, category.ID).WithContext(ctx).Exec()

	return err
}

func (r *CategoryRepositoryCassandra) Remove(ctx context.Context, id string) error {
	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	ctxI, _ := context.WithCancel(ctx)
	category, err := r.FetchByID(ctxI, id, true)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	err = s.Query(`UPDATE alexa1.category SET active = ? WHERE external_id = ? AND id = ?`, false,
		category.ExternalID, category.ID).WithContext(ctx).Exec()
	if err != nil && err == gocql.ErrNotFound {
		return exception.EntityNotFound
	}

	return err
}

func (r *CategoryRepositoryCassandra) Restore(ctx context.Context, id string) error {
	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	ctxI, _ := context.WithCancel(ctx)
	category, err := r.FetchByID(ctxI, id, false)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	err = s.Query(`UPDATE alexa1.category SET active = ? WHERE external_id = ? AND id = ?`, true,
		category.ExternalID, category.ID).WithContext(ctx).Exec()

	return err
}

func (r *CategoryRepositoryCassandra) HardRemove(ctx context.Context, id string) error {
	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	ctxI, _ := context.WithCancel(ctx)
	category, err := r.FetchByID(ctxI, id, false)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	err = s.Query("DELETE FROM alexa1.category WHERE external_id = ? AND id = ?", category.ExternalID, category.ID).WithContext(ctx).Exec()

	return err
}

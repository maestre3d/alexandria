package infrastructure

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/gocql/gocql"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"sync"
)

type CategoryRootCassandraRepository struct {
	mu   *sync.RWMutex
	pool *gocql.ClusterConfig
}

func NewCategoryRootCassandraRepository(pool *gocql.ClusterConfig) *CategoryRootCassandraRepository {
	return &CategoryRootCassandraRepository{
		mu:   new(sync.RWMutex),
		pool: pool,
	}
}

func (r *CategoryRootCassandraRepository) Save(ctx context.Context, root domain.CategoryByRoot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Query(`INSERT INTO alexa1.category_by_root (root_id, category) VALUES 
		(?, ?)`, root.RootID, root.CategoryList).WithContext(ctx).Exec()
}

func (r *CategoryRootCassandraRepository) AddItem(ctx context.Context, rootID string, item map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Query(`UPDATE alexa1.category_by_root SET category = category + ? WHERE root_id = ?`,
		item, rootID).WithContext(ctx).Exec()
}

func (r *CategoryRootCassandraRepository) FetchByRoot(ctx context.Context, rootID string) (*domain.CategoryByRoot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	categoryRoot := new(domain.CategoryByRoot)
	err = s.Query(`SELECT * FROM alexa1.category_by_root WHERE TOKEN(root_id) = TOKEN(?) LIMIT 1`, rootID).Consistency(gocql.One).WithContext(ctx).
		Scan(&categoryRoot.RootID, &categoryRoot.CategoryList)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, exception.EntityNotFound
		}
		return nil, err
	}

	return categoryRoot, nil
}

func (r *CategoryRootCassandraRepository) Fetch(ctx context.Context, params core.PaginationParams) ([]*domain.CategoryByRoot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	iter := s.Query(`SELECT * FROM alexa1.category_by_root WHERE TOKEN(root_id) >= TOKEN(?) LIMIT ?`, params.Token, params.Size).
		WithContext(ctx).PageSize(params.Size).Iter()
	if iter.NumRows() == 0 {
		return nil, exception.EntitiesNotFound
	}

	categories := make([]*domain.CategoryByRoot, 0)
	iterCat := new(domain.CategoryByRoot)
	for iter.Scan(&iterCat.RootID, &iterCat.CategoryList) {
		categoryRoot := new(domain.CategoryByRoot)
		categoryRoot = iterCat
		categories = append(categories, categoryRoot)
	}

	return categories, nil
}

func (r *CategoryRootCassandraRepository) RemoveItem(ctx context.Context, rootID, categoryID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Query(`DELETE category[?] FROM alexa1.category_by_root WHERE root_id = ?`, categoryID, rootID).WithContext(ctx).Exec()
}

func (r *CategoryRootCassandraRepository) HardRemoveList(ctx context.Context, rootID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Query(`DELETE FROM alexa1.category_by_root WHERE root_id = ?`, rootID).WithContext(ctx).Exec()
}

package infrastructure

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-redis/redis/v7"
	"github.com/gocql/gocql"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"sync"
	"time"
)

var (
	latencyMs    = stats.Float64("category/cassandra/latency", "The latency in milliseconds per Apache Cassandra read", "ms")
	keyMethod, _ = tag.NewKey("method")
	keyStatus, _ = tag.NewKey("status")
	keyError, _  = tag.NewKey("error")
)

type CategoryRepositoryCassandra struct {
	mu      *sync.RWMutex
	pool    *gocql.ClusterConfig
	memPool *redis.Client
}

func NewCategoryRepositoryCassandra(pool *gocql.ClusterConfig, memPool *redis.Client) *CategoryRepositoryCassandra {
	return &CategoryRepositoryCassandra{
		mu:      new(sync.RWMutex),
		pool:    pool,
		memPool: memPool,
	}
}

func (r *CategoryRepositoryCassandra) Save(ctx context.Context, category domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	startTime := time.Now()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	// Enable dist. tracing
	ctxT, span := trace.StartSpan(ctx, "category_save")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "write row in cassandra",
	})
	span.AddAttributes(trace.StringAttribute("operation", "save"), trace.StringAttribute("db.driver", "cassandra"))

	// Enable census metrics
	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.save"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM
	defer func() {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(startTime).Nanoseconds())/1e6))
	}()

	err = s.Query(`INSERT INTO alexa1.category (id, external_id, category_name, create_time, update_time, active) VALUES  
		(?, ?, ?, ?, ?, ?)`, gocql.TimeUUID(), category.ExternalID, category.Name, category.CreateTime, category.UpdateTime, category.Active).
		WithContext(ctx).Exec()

	return err
}

func (r *CategoryRepositoryCassandra) FetchByID(ctx context.Context, id string, activeOnly bool) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	startTime := time.Now()

	// _, id = DecodeCassandraID(id)

	// TODO: Redis cache-aside pattern impl. before Cassandra read op.

	s, err := r.pool.CreateSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	// Enable dist. tracing
	ctxT, span := trace.StartSpan(ctx, "category_fetch_by_id")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "read row in cassandra",
	})
	span.AddAttributes(trace.StringAttribute("operation", "fetch_by_id"), trace.StringAttribute("db.driver", "cassandra"))

	// Enable census metrics
	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.fetch_by_id"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return nil, err
	}
	ctx = ctxM
	defer func() {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(startTime).Nanoseconds())/1e6))
	}()

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
	// Encode internal URL
	// category.ExternalID = EncodeCassandraID(category.ID, category.ExternalID)

	return category, nil
}

func (r *CategoryRepositoryCassandra) Fetch(ctx context.Context, params core.PaginationParams, filter core.FilterParams) ([]*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	startTime := time.Now()

	s, err := r.pool.CreateSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	// Enable dist. tracing
	ctxT, span := trace.StartSpan(ctx, "category_fetch")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "read row in cassandra",
	})
	span.AddAttributes(trace.StringAttribute("operation", "fetch"), trace.StringAttribute("db.driver", "cassandra"))

	// Enable census metrics
	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.fetch"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return nil, err
	}
	ctx = ctxM
	defer func() {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(startTime).Nanoseconds())/1e6))
	}()

	builder := &CategoryCassandraBuilder{Statement: `SELECT * FROM alexa1.category WHERE `}
	for k, v := range filter {
		switch {
		case k == "query" && v != "":
			builder.Name(v).And()
			continue
		}
	}

	if params.Token != "" {
		// _, extID := DecodeCassandraID(params.Token)
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
		// category.ExternalID = EncodeCassandraID(category.ID, category.ExternalID)
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
	startTime := time.Now()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	// Enable dist. tracing
	ctxT, span := trace.StartSpan(ctx, "category_replace")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "write row in cassandra",
	})
	span.AddAttributes(trace.StringAttribute("operation", "replace"), trace.StringAttribute("db.driver", "cassandra"))

	// Enable census metrics
	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.replace"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM
	defer func() {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(startTime).Nanoseconds())/1e6))
	}()

	err = s.Query(`UPDATE alexa1.category SET category_name = ?, update_time = ? WHERE external_id = ? AND id = ?`, category.Name,
		category.UpdateTime, category.ExternalID, category.ID).WithContext(ctx).Exec()

	return err
}

func (r *CategoryRepositoryCassandra) Remove(ctx context.Context, id string) error {
	startTime := time.Now()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	// Enable dist. tracing
	ctxT, span := trace.StartSpan(ctx, "category_remove")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "write row in cassandra",
	})
	span.AddAttributes(trace.StringAttribute("operation", "remove"), trace.StringAttribute("db.driver", "cassandra"))

	// Enable census metrics
	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.remove"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM
	defer func() {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(startTime).Nanoseconds())/1e6))
	}()
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
	startTime := time.Now()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	// Enable dist. tracing
	ctxT, span := trace.StartSpan(ctx, "category_restore")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "write row in cassandra",
	})
	span.AddAttributes(trace.StringAttribute("operation", "restore"), trace.StringAttribute("db.driver", "cassandra"))

	// Enable census metrics
	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.restore"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM
	defer func() {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(startTime).Nanoseconds())/1e6))
	}()
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
	startTime := time.Now()

	s, err := r.pool.CreateSession()
	if err != nil {
		return err
	}
	defer s.Close()

	// Enable dist. tracing
	ctxT, span := trace.StartSpan(ctx, "category_hard_remove")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "delete row in cassandra",
	})
	span.AddAttributes(trace.StringAttribute("operation", "hard_remove"), trace.StringAttribute("db.driver", "cassandra"))

	// Enable census metrics
	ctxM, err := tag.New(ctx, tag.Insert(keyMethod, "category.hard_remove"), tag.Insert(keyStatus, "OK"))
	if err != nil {
		return err
	}
	ctx = ctxM
	defer func() {
		if err != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(keyStatus, "ERROR"), tag.Insert(keyError, err.Error()))
		}

		stats.Record(ctx, latencyMs.M(float64(time.Since(startTime).Nanoseconds())/1e6))
	}()
	ctxI, _ := context.WithCancel(ctx)
	category, err := r.FetchByID(ctxI, id, false)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	err = s.Query("DELETE FROM alexa1.category WHERE external_id = ? AND id = ?", category.ExternalID, category.ID).WithContext(ctx).Exec()
	// TODO: Catch 404 errors

	return err
}

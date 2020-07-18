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
	mu      *sync.Mutex
	pool    *gocql.ClusterConfig
	memPool *redis.Client
}

func NewCategoryRepositoryCassandra(pool *gocql.ClusterConfig, memPool *redis.Client) *CategoryRepositoryCassandra {
	return &CategoryRepositoryCassandra{
		mu:      new(sync.Mutex),
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
		Message: "store row in cassandra",
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

func (r *CategoryRepositoryCassandra) FetchByID(ctx context.Context, id string) (*domain.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	startTime := time.Now()

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

	category := new(domain.Category)
	err = s.Query(`SELECT * FROM alexa1.category WHERE external_id = ? AND active = TRUE LIMIT 1 ALLOW FILTERING`,
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
	r.mu.Lock()
	defer r.mu.Unlock()
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
		builder.Statement += `external_id = ` + params.Token
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

	category := new(domain.Category)
	categories := make([]*domain.Category, 0)
	for iter.Scan(&category.ExternalID, &category.ID, &category.Active, &category.Name, &category.CreateTime, &category.UpdateTime) {
		categories = append(categories, category)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoryRepositoryCassandra) Replace(ctx context.Context, category domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}

func (r *CategoryRepositoryCassandra) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}

func (r *CategoryRepositoryCassandra) Restore(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}

func (r *CategoryRepositoryCassandra) HardRemove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}

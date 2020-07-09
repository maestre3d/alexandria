package infrastructure

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/persistence"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	_ "gocloud.dev/blob/s3blob"
	"gocloud.dev/gcerrors"
	"sync"
)

type BlobDynamoRepository struct {
	cfg    *config.Kernel
	logger log.Logger
	mu     *sync.Mutex
}

func NewBlobDynamoRepository(logger log.Logger, cfg *config.Kernel) *BlobDynamoRepository {
	return &BlobDynamoRepository{
		cfg:    cfg,
		logger: logger,
		mu:     new(sync.Mutex),
	}
}

func (r *BlobDynamoRepository) Save(ctx context.Context, blobRef domain.Blob) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	coll, _, err := persistence.NewDynamoDBCollectionPool(ctx, r.cfg)
	if err != nil {
		return err
	}
	defer coll.Close()

	return coll.Put(ctx, &blobRef)
}

func (r *BlobDynamoRepository) FetchByID(ctx context.Context, id string) (*domain.Blob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	coll, _, err := persistence.NewDynamoDBCollectionPool(ctx, r.cfg)
	if err != nil {
		return nil, err
	}
	defer coll.Close()

	b := &domain.Blob{ID: id}
	err = coll.Get(ctx, b)
	if err != nil {
		if gcerrors.Code(err) == gcerrors.NotFound {
			return nil, exception.EntityNotFound
		}
		return nil, err
	}

	return b, nil
}

func (r *BlobDynamoRepository) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	coll, _, err := persistence.NewDynamoDBCollectionPool(ctx, r.cfg)
	if err != nil {
		return err
	}
	defer coll.Close()

	b := &domain.Blob{ID: id}
	return coll.Delete(ctx, b)
}

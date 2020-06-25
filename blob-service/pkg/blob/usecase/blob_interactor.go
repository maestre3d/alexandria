package usecase

import (
	"context"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
)

type BlobInteractor interface {
	Store(ctx context.Context, ag *domain.BlobAggregate) (*domain.Blob, error)
	Get(ctx context.Context, id, service string) (*domain.Blob, error)
	Delete(ctx context.Context, id, service string) error
}

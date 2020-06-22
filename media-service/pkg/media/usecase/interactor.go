package usecase

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
)

type MediaInteractor interface {
	Create(ctx context.Context, aggregate *domain.MediaAggregate) (*domain.Media, error)
	List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) ([]*domain.Media, string, error)
	Get(ctx context.Context, id string) (*domain.Media, error)
	Update(ctx context.Context, aggregate *domain.MediaUpdateAggregate) (*domain.Media, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

type MediaSAGAInteractor interface {
	VerifyAuthor(ctx context.Context, rootID string) error
	Done(ctx context.Context, rootID, operation string) error
	Failed(ctx context.Context, rootID, operation, backup string) error
}

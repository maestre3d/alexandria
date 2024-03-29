package usecase

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
)

type AuthorInteractor interface {
	Create(ctx context.Context, aggregate *domain.AuthorAggregate) (*domain.Author, error)
	List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) ([]*domain.Author, string, error)
	Get(ctx context.Context, id string) (*domain.Author, error)
	Update(ctx context.Context, aggregate *domain.AuthorUpdateAggregate) (*domain.Author, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

type AuthorSAGAInteractor interface {
	Verify(ctx context.Context, service string, authorsJSON []byte) error
	Done(ctx context.Context, rootID, operation string) error
	Failed(ctx context.Context, rootID, operation, backup string) error
	UpdatePicture(ctx context.Context, rootID string, urlJSON []byte) error
	RemovePicture(ctx context.Context, rootID []byte) error
}

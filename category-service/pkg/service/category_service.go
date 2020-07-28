package service

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
)

type Category interface {
	Create(ctx context.Context, name string) (*domain.Category, error)
	Get(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context, token, limit string, filter core.FilterParams) ([]*domain.Category, string, error)
	Update(ctx context.Context, id string, name string) (*domain.Category, error)
	Delete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

type CategoryRoot interface {
	CreateList(ctx context.Context, categoryID, rootID string) (*domain.CategoryByRoot, error)
	Add(ctx context.Context, categoryID, rootID string) error
	GetByRoot(ctx context.Context, rootID string) (*domain.CategoryByRoot, error)
	List(ctx context.Context, token, limit string) ([]*domain.CategoryByRoot, string, error)
	DeleteItem(ctx context.Context, categoryID string) error
	DeleteList(ctx context.Context, rootID string) error
}

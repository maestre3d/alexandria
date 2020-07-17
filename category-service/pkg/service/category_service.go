package service

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
)

type Category interface {
	Create(ctx context.Context, name, service string) (*domain.Category, error)
	Get(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context, nextToken, limit string, filter core.FilterParams) ([]*domain.Category, string, error)
	Update(ctx context.Context, id string, name string) (*domain.Category, error)
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

type CategoryRelation interface {
	Add(ctx context.Context, id, rootID, service string) error
	List(ctx context.Context, rootID string)
	Remove(ctx context.Context, id, rootID, service string) error
}

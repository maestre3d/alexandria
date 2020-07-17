package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
)

type CategoryUseCase struct {
	logger log.Logger
	repo   domain.CategoryRepository
	event  domain.CategoryEventBus
}

func NewCategoryUseCase(logger log.Logger, repo domain.CategoryRepository, event domain.CategoryEventBus) *CategoryUseCase {
	return &CategoryUseCase{
		logger: logger,
		repo:   repo,
		event:  event,
	}
}

func (u *CategoryUseCase) Create(ctx context.Context, name, service string) (*domain.Category, error) {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	category := domain.NewCategory(name)
	if err := category.IsValid(); err != nil {
		return nil, err
	}

	err := u.repo.Save(ctxI, *category)
	if err != nil {
		return nil, err
	}

	errC := make(chan error)
	go func() {
		err = u.event.StartCreate(ctxI, *category)
		if err != nil {
			_ = level.Error(u.logger).Log(err)
			err = u.event.HardRemoved(ctx, category.ExternalID)
			if err != nil {
				_ = level.Error(u.logger).Log(err)
			}
		}

		_ = level.Info(u.logger).Log(fmt.Sprintf("event %s sent", domain.EntityVerify(service)))
	}()

	select {
	case err = <-errC:
		if err != nil {
			return nil, err
		}
		break
	}

	return category, nil
}

func (u *CategoryUseCase) Get(ctx context.Context, id string) (*domain.Category, error) {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	// We could add total_views
	return u.repo.FetchByID(ctxI, id)
}

func (u *CategoryUseCase) List(ctx context.Context, token, limit string, filter core.FilterParams) ([]*domain.Category, string, error) {
	params := core.NewPaginationParams(token, limit)
	nextToken := ""

	// Increment to get nextToken
	params.Size++
	categories, err := u.repo.Fetch(ctx, *params, filter)
	if err != nil {
		return nil, "", err
	}

	if len(categories) > 1 {
		nextToken = categories[len(categories)-1].ExternalID
		categories = categories[:len(categories)-2]
	}

	return categories, nextToken, nil
}

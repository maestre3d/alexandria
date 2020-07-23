package interactor

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"strings"
	"time"
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

func (u *CategoryUseCase) Create(ctx context.Context, name string) (*domain.Category, error) {
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
		err = u.event.Created(ctxI, *category)
		if err != nil {
			// Rollback
			errC <- u.event.HardRemoved(ctxI, category.ExternalID)
			return
		}
		errC <- nil
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
	return u.repo.FetchByID(ctxI, id, true)
}

func (u *CategoryUseCase) List(ctx context.Context, token, limit string, filter core.FilterParams) ([]*domain.Category, string, error) {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	params := core.NewPaginationParams(token, limit)
	nextToken := ""

	// Increment to get nextToken
	params.Size++
	categories, err := u.repo.Fetch(ctxI, *params, filter)
	if err != nil {
		return nil, "", err
	}

	if len(categories) >= params.Size {
		nextToken = categories[len(categories)-1].ExternalID
		categories = categories[0 : len(categories)-1]
	}

	return categories, nextToken, nil
}

func (u *CategoryUseCase) Update(ctx context.Context, id, name string) (*domain.Category, error) {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	// Non-atomic update
	category, err := u.Get(ctxI, id)
	if err != nil {
		return nil, err
	}
	snapshot := category

	if name != "" {
		category.Name = strings.Title(name)
	}
	category.UpdateTime = time.Now()

	err = u.repo.Replace(ctxI, *category)
	if err != nil {
		return nil, err
	}

	errC := make(chan error)
	go func() {
		err = u.event.Updated(ctxI, *category)
		if err != nil {
			// Rollback
			errC <- u.repo.Replace(ctxI, *snapshot)
			return
		}

		errC <- nil
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

func (u *CategoryUseCase) Delete(ctx context.Context, id string) error {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	err := u.repo.Remove(ctxI, id)
	if err != nil {
		return err
	}

	errC := make(chan error)
	go func() {
		err = u.event.Removed(ctxI, id)
		if err != nil {
			// Rollback
			errC <- u.repo.Restore(ctxI, id)
			return
		}

		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (u *CategoryUseCase) Restore(ctx context.Context, id string) error {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	err := u.repo.Restore(ctxI, id)
	if err != nil {
		return err
	}

	errC := make(chan error)
	go func() {
		err = u.event.Restored(ctxI, id)
		if err != nil {
			// Rollback
			errC <- u.repo.Remove(ctxI, id)
			return
		}

		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (u *CategoryUseCase) HardDelete(ctx context.Context, id string) error {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	snapshot, err := u.repo.FetchByID(ctxI, id, false)
	if err != nil {
		return err
	}

	err = u.repo.HardRemove(ctxI, id)
	if err != nil {
		return err
	}

	errC := make(chan error)
	go func() {
		err = u.event.HardRemoved(ctxI, id)
		if err != nil {
			// Rollback
			errC <- u.repo.Save(ctxI, *snapshot)
			return
		}

		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
		break
	}

	return nil
}

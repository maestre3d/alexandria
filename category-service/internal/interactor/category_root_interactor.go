package interactor

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
)

type CategoryRootUseCase struct {
	logger       log.Logger
	repo         domain.CategoryRootRepository
	categoryRepo domain.CategoryRepository
	// eventBus     domain.CategoryRootEventBus
}

func NewCategoryRootUseCase(logger log.Logger, repo domain.CategoryRootRepository,
	categoryRepo domain.CategoryRepository) *CategoryRootUseCase {
	return &CategoryRootUseCase{
		logger:       logger,
		repo:         repo,
		categoryRepo: categoryRepo,
		//eventBus:     event,
	}
}

func (u *CategoryRootUseCase) CreateList(ctx context.Context, categoryID, rootID string) (*domain.CategoryByRoot, error) {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	// Get normalized category
	ctxR, _ := context.WithCancel(ctx)
	category, err := u.categoryRepo.FetchByID(ctxR, categoryID, true)
	if err != nil {
		return nil, err
	}

	// Use internal ID for ref keys/denormalized CF
	categoryRoot := domain.NewCategoryByRoot(rootID, category.ExternalID, category.Name)

	// TODO: Add event propagation/transactions
	err = u.repo.Save(ctxI, *categoryRoot)
	if err != nil {
		return nil, err
	}

	return categoryRoot, nil
}

func (u *CategoryRootUseCase) Add(ctx context.Context, categoryID, rootID string) error {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	// Get normalized category
	ctxR, _ := context.WithCancel(ctx)
	category, err := u.categoryRepo.FetchByID(ctxR, categoryID, true)
	if err != nil {
		return err
	}

	// Use internal ID for ref keys/denormalized CF
	list := map[string]string{category.ExternalID: category.Name}

	// TODO: Add event propagation/transactions
	return u.repo.AddItem(ctxI, rootID, list)
}

func (u *CategoryRootUseCase) GetByRoot(ctx context.Context, rootID string) (*domain.CategoryByRoot, error) {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	return u.repo.FetchByRoot(ctxI, rootID)
}

func (u *CategoryRootUseCase) List(ctx context.Context, token, limit string) ([]*domain.CategoryByRoot, string, error) {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	params := core.NewPaginationParams(token, limit)
	nextToken := ""

	// Increment to get nextToken
	params.Size++
	categories, err := u.repo.Fetch(ctxI, *params)
	if err != nil {
		return nil, "", err
	}

	if len(categories) >= params.Size {
		nextToken = categories[len(categories)-1].RootID
		categories = categories[0 : len(categories)-1]
	}

	return categories, nextToken, nil
}

func (u *CategoryRootUseCase) DeleteItem(ctx context.Context, rootID, categoryID string) error {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	return u.repo.RemoveItem(ctxI, rootID, categoryID)
}

func (u *CategoryRootUseCase) DeleteList(ctx context.Context, rootID string) error {
	ctxI, cancel := context.WithCancel(ctx)
	defer cancel()

	return u.repo.HardRemoveList(ctxI, rootID)
}

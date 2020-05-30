package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
)

// AuthorUseCase Author interact actions
type AuthorUseCase struct {
	log        log.Logger
	repository domain.IAuthorRepository
	eventBus   domain.IAuthorEventBus
}

// NewAuthorUseCase Create a new author interact
func NewAuthorUseCase(logger log.Logger, repository domain.IAuthorRepository, bus domain.IAuthorEventBus) *AuthorUseCase {
	return &AuthorUseCase{logger, repository, bus}
}

// Create Store a new entity
func (u *AuthorUseCase) Create(ctx context.Context, aggregate *domain.AuthorAggregate) (*domain.Author, error) {
	// Validate non-string values
	birth, err := time.Parse(core.RFC3339Micro, aggregate.BirthDate)
	if err != nil {
		return nil, exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"birth_date", core.RFC3339Micro))
	}

	author := domain.NewAuthor(aggregate.FirstName, aggregate.LastName, aggregate.DisplayName, aggregate.OwnerID, birth)
	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Save(ctx, author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	go func() {
		err = u.eventBus.Created(ctx, author)
		if err != nil {
			_ = u.log.Log("method", "author.interactor.create", "err", err.Error())
		} else {
			_ = u.log.Log("method", "author.interactor.create", "msg", "ALEXANDRIA_AUTHOR_CREATED event published")
		}
	}()

	return author, nil
}

// List Get an author's list
func (u *AuthorUseCase) List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) (output []*domain.Author, nextToken string, err error) {
	params := core.NewPaginationParams(pageToken, pageSize)
	params.Size++
	output, err = u.repository.Fetch(ctx, params, filterParams)

	nextToken = ""
	if len(output) >= params.Size {
		nextToken = output[len(output)-1].ExternalID
		output = output[0 : len(output)-1]
	}
	return
}

// Get Obtain one author
func (u *AuthorUseCase) Get(ctx context.Context, id string) (*domain.Author, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, exception.InvalidID
	}

	return u.repository.FetchByID(ctx, id)
}

// Update Update an author dynamically
func (u *AuthorUseCase) Update(ctx context.Context, id string, status string, aggregate *domain.AuthorAggregate) (*domain.Author, error) {
	// Check if body has values, if not return to avoid any transaction
	if aggregate.FirstName == "" && aggregate.LastName == "" && aggregate.DisplayName == "" && aggregate.BirthDate == "" &&
		aggregate.OwnerID == "" && status == "" {
		return nil, exception.EmptyBody
	}

	// Get previous version
	author, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update entity dynamically
	if aggregate.BirthDate != "" {
		birth, err := time.Parse(core.RFC3339Micro, aggregate.BirthDate)
		if err != nil {
			return nil, exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"birth_date", core.RFC3339Micro))
		}
		author.BirthDate = birth
	}
	if aggregate.FirstName != "" {
		author.FirstName = aggregate.FirstName
	}
	if aggregate.LastName != "" {
		author.LastName = aggregate.LastName
	}
	if aggregate.DisplayName != "" {
		author.DisplayName = aggregate.DisplayName
	}
	if status != "" {
		author.Status = status
	}

	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Replace(ctx, author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Fire up "alexandria.author.updated" domain event

	return author, nil
}

// Delete Remove an author from the store
func (u *AuthorUseCase) Delete(ctx context.Context, id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidID
	}

	err = u.repository.Remove(ctx, id)
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	go func() {
		if err == nil {
			if errBus := u.eventBus.Deleted(ctx, id, false); errBus != nil {
				_ = u.log.Log("method", "author.interactor.delete", "err", errBus.Error())
			} else {
				_ = u.log.Log("method", "author.interactor.delete", "msg", "ALEXANDRIA_AUTHOR_DELETED event published")
			}
		}
	}()

	return err
}

// Restore recover an author from the store
func (u *AuthorUseCase) Restore(ctx context.Context, id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidID
	}

	err = u.repository.Restore(ctx, id)
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	go func() {
		if err == nil {
			if errBus := u.eventBus.Restored(ctx, id); errBus != nil {
				_ = u.log.Log("method", "author.interactor.restore", "err", errBus.Error())
			} else {
				_ = u.log.Log("method", "author.interactor.restore", "msg", "ALEXANDRIA_AUTHOR_DELETED event published")
			}
		}
	}()

	return err
}

// HardDelete Remove an author from the store permanently
func (u *AuthorUseCase) HardDelete(ctx context.Context, id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidID
	}

	err = u.repository.HardRemove(ctx, id)
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	go func() {
		if err == nil {
			if errBus := u.eventBus.Deleted(ctx, id, true); errBus != nil {
				_ = u.log.Log("method", "author.interactor.harddelete", "err", errBus.Error())
			} else {
				_ = u.log.Log("method", "author.interactor.harddelete", "msg", "ALEXANDRIA_AUTHOR_DELETED event published")
			}
		}
	}()

	return err
}

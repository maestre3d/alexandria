package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"strconv"
	"time"
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
	author := domain.NewAuthor(aggregate.FirstName, aggregate.LastName, aggregate.DisplayName, aggregate.OwnershipType,
		domain.NewOwner(aggregate.OwnerID, string(domain.RootRole)))
	err := author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Save(ctx, author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// Transaction/interaction event, required validation, use concurrent-safe routine
	errChan := make(chan error)
	go func() {
		err = u.eventBus.Created(ctx, author)
		if err != nil {
			_ = u.log.Log("method", "author.interactor.create", "err", err.Error())

			// Rollback
			err = u.repository.HardRemove(ctx, author.ExternalID)
			if err != nil {
				_ = u.log.Log("method", "author.interactor.create", "err", err.Error())
			}

			_ = u.log.Log("method", "author.interactor.create", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.create", "msg", "ALEXANDRIA_AUTHOR_CREATED event published")
		}

		errChan <- err
	}()

	select {
	case err = <-errChan:
		if err != nil {
			return nil, err
		}
	}

	return author, err
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
	author, err := u.repository.FetchByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if author != nil {
		author.TotalViews++
		// Using repo directly to avoid unused fields on usecase's update
		err = u.repository.Replace(ctx, author)
		if err != nil {
			_ = u.log.Log("method", "author.interactor.get", "msg", fmt.Sprintf("could not update total_views for author %s, error: %s",
				author.ExternalID, err.Error()))
		}
	}

	return author, nil
}

// Update Update an author dynamically
func (u *AuthorUseCase) Update(ctx context.Context, aggregate *domain.AuthorUpdateAggregate) (*domain.Author, error) {
	// Check if body has values, if not return to avoid any transaction
	if aggregate.RootAggregate.FirstName == "" && aggregate.RootAggregate.LastName == "" && aggregate.RootAggregate.DisplayName == "" &&
		aggregate.RootAggregate.OwnershipType == "" && aggregate.RootAggregate.OwnerID == "" &&
		len(aggregate.Owners) == 0 {
		return nil, exception.EmptyBody
	}

	// Get previous version
	// Using repository directly to avoid non-organic total_views
	author, err := u.repository.FetchByID(ctx, aggregate.ID)
	authorBackup := author
	if err != nil {
		return nil, err
	}

	// Update entity dynamically
	if aggregate.RootAggregate.FirstName != "" {
		author.FirstName = aggregate.RootAggregate.FirstName
	}
	if aggregate.RootAggregate.LastName != "" {
		author.LastName = aggregate.RootAggregate.LastName
	}
	if aggregate.RootAggregate.DisplayName != "" {
		author.DisplayName = aggregate.RootAggregate.DisplayName
	}
	if aggregate.RootAggregate.OwnershipType != "" {
		author.OwnershipType = aggregate.RootAggregate.OwnershipType
	}
	if aggregate.Verified != "" {
		verified, err := strconv.ParseBool(aggregate.Verified)
		if err != nil {
			return nil, exception.NewErrorDescription(exception.InvalidFieldFormat,
				fmt.Sprintf(exception.InvalidFieldFormatString, "verified", "boolean"))
		}

		author.Verified = verified
	}
	if aggregate.Picture != "" {
		author.Picture = &aggregate.Picture
	}
	// Update owner id
	if aggregate.RootAggregate.OwnerID != "" {
		for _, owner := range author.Owners {
			if owner.Role == string(domain.RootRole) {
				owner.ID = aggregate.RootAggregate.OwnerID
				break
			}
		}
	}

	newOwners := make([]*domain.Owner, 0)
	rootCount := 0
	// Add new owners
	for _, owner := range aggregate.Owners {
		if owner.Role == string(domain.RootRole) {
			rootCount++
		}

		// Avoid multiple root owners
		if rootCount > 1 {
			return nil, exception.NewErrorDescription(exception.InvalidFieldFormat,
				fmt.Sprintf(exception.InvalidFieldFormatString, "owner_role", "only one user with role owner"))
		}

		newOwners = append(newOwners, &domain.Owner{
			ID:   owner.ID,
			Role: owner.Role,
		})
	}
	// Avoid non-owner user
	if rootCount == 0 {
		return nil, exception.NewErrorDescription(exception.InvalidFieldFormat,
			fmt.Sprintf(exception.InvalidFieldFormatString, "owner_role", "one user with role owner"))
	}

	// Avoid more than 15 owners
	if len(newOwners) > 15 {
		return nil, exception.NewErrorDescription(exception.InvalidFieldRange,
			fmt.Sprintf(exception.InvalidFieldRangeString, "owners", "1", "15"))
	}

	author.Owners = newOwners
	author.UpdateTime = time.Now()

	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Replace(ctx, author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	errChan := make(chan error)
	go func() {
		err := u.eventBus.Updated(ctx, author.Owners)
		if err != nil {
			_ = u.log.Log("method", "author.interactor.update", "err", err.Error())

			// Rollback
			err = u.repository.Replace(ctx, authorBackup)
			if err != nil {
				_ = u.log.Log("method", "author.interactor.update", "err", err.Error())
			}

			_ = u.log.Log("method", "author.interactor.update", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.update", "msg", "ALEXANDRIA_AUTHOR_UPDATED event published")
		}

		errChan <- err
	}()

	select {
	case err = <-errChan:
		if err != nil {
			return nil, err
		}
	}

	return author, nil
}

// Delete Remove an author from the store
func (u *AuthorUseCase) Delete(ctx context.Context, id string) error {
	err := u.repository.Remove(ctx, id)
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Move domain events and misc to infrastructure layer, use SQL transactions to handle operations atomically
	go func() {
		if err == nil {
			ctxTime, cancelCtx := context.WithTimeout(ctx, time.Second*10)
			defer cancelCtx()
			if errBus := u.eventBus.Deleted(ctxTime, id, false); errBus != nil {
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
	err := u.repository.Restore(ctx, id)
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
	err := u.repository.HardRemove(ctx, id)
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	go func() {
		if err == nil {
			ctxTime, cancelCtx := context.WithTimeout(ctx, time.Second*10)
			defer cancelCtx()
			if errBus := u.eventBus.Deleted(ctxTime, id, true); errBus != nil {
				_ = u.log.Log("method", "author.interactor.harddelete", "err", errBus.Error())
			} else {
				_ = u.log.Log("method", "author.interactor.harddelete", "msg", "ALEXANDRIA_AUTHOR_DELETED event published")
			}
		}
	}()

	return err
}

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

// Author
// business actions
type Author struct {
	log        log.Logger
	repository domain.AuthorRepository
	event      domain.AuthorEventBus
}

// NewAuthor Create a new author interact
func NewAuthor(logger log.Logger, repository domain.AuthorRepository, bus domain.AuthorEventBus) *Author {
	return &Author{logger, repository, bus}
}

// Create Store a new entity
func (u *Author) Create(ctx context.Context, aggregate *domain.AuthorAggregate) (*domain.Author, error) {
	author := domain.NewAuthor(aggregate.FirstName, aggregate.LastName, aggregate.DisplayName, aggregate.OwnershipType, aggregate.OwnerID,
		aggregate.Country)
	err := author.IsValid()
	if err != nil {
		return nil, err
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err = u.repository.Save(ctxR, *author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// Transaction/interaction event, required owner/user validation, use concurrent-safe routine
	errChan := make(chan error)
	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()

		err = u.event.StartCreate(ctxE, *author)
		if err != nil {
			_ = u.log.Log("method", "author.interactor.create", "err", err.Error())

			// Rollback
			err = u.repository.HardRemove(ctxE, author.ExternalID)
			if err != nil {
				_ = u.log.Log("method", "author.interactor.create", "err", err.Error())
			}

			_ = u.log.Log("method", "author.interactor.create", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.create", "msg", domain.OwnerVerify+" integration event published")
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
func (u *Author) List(ctx context.Context, pageToken, pageSize string, filterParams core.FilterParams) (output []*domain.Author, nextToken string, err error) {
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	params := core.NewPaginationParams(pageToken, pageSize)
	params.Size++
	output, err = u.repository.Fetch(ctxR, *params, filterParams)

	nextToken = ""
	if len(output) >= params.Size {
		nextToken = output[len(output)-1].ExternalID
		output = output[0 : len(output)-1]
	}
	return
}

// Get Obtain one author
func (u *Author) Get(ctx context.Context, id string) (*domain.Author, error) {
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	author, err := u.repository.FetchByID(ctxR, id, false)
	if err != nil {
		return nil, err
	}

	if author != nil {
		author.TotalViews++
		// Using repo directly to avoid unused fields on use case's update
		err = u.repository.Replace(ctx, *author)
		if err != nil {
			_ = u.log.Log("method", "author.interactor.get", "msg", fmt.Sprintf("could not update total_views for author %s, error: %s",
				author.ExternalID, err.Error()))
		}
	}

	return author, nil
}

// Update Update an author dynamically
func (u *Author) Update(ctx context.Context, aggregate *domain.AuthorUpdateAggregate) (*domain.Author, error) {
	// Check if body has values, if not return to avoid any transaction
	if aggregate.RootAggregate.FirstName == "" && aggregate.RootAggregate.LastName == "" && aggregate.RootAggregate.DisplayName == "" &&
		aggregate.RootAggregate.OwnershipType == "" && aggregate.RootAggregate.OwnerID == "" {
		return nil, exception.EmptyBody
	}

	// Get previous version
	// Using repository directly to avoid non-organic total_views increment
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	author, err := u.repository.FetchByID(ctxR, aggregate.ID, false)
	if err != nil {
		return nil, err
	}
	authorBackup := author

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
	// If new owner id was given, then set author state to pending to start proper
	// transaction
	if aggregate.RootAggregate.OwnerID != "" {
		author.OwnerID = aggregate.RootAggregate.OwnerID
		author.Status = domain.StatusPending
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
	if aggregate.RootAggregate.Country != "" {
		author.Country = aggregate.RootAggregate.Country
	}

	author.UpdateTime = time.Now()

	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Replace(ctxR, *author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// Transaction/interaction event, required owner/user validation, use concurrent-safe routine
	errChan := make(chan error)
	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()

		// If author owner id is changed, then start transaction with integration event, if not
		// send a simple domain event to propagate side-effects
		var eventStr string
		if author.Status == domain.StatusPending {
			err = u.event.StartUpdate(ctxE, *author, *authorBackup)
			if err == nil {
				eventStr = domain.OwnerVerify + " event published"
			}
		} else {
			err = u.event.Updated(ctxE, *author)
			if err == nil {
				eventStr = domain.AuthorUpdated + " event published"
			}
		}

		if err != nil {
			_ = u.log.Log("method", "author.interactor.update", "err", err.Error())

			// Rollback
			err = u.repository.Replace(ctxE, *authorBackup)
			if err != nil {
				_ = u.log.Log("method", "author.interactor.update", "err", err.Error())
			}

			_ = u.log.Log("method", "author.interactor.update", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.update", "msg", eventStr)
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
func (u *Author) Delete(ctx context.Context, id string) error {
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err := u.repository.Remove(ctxR, id)
	if err != nil {
		return err
	}
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// Propagate side-effects
	errC := make(chan error)
	defer close(errC)
	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()
		if err = u.event.Removed(ctxE, id); err != nil {
			_ = u.log.Log("method", "author.interactor.delete", "err", err.Error())

			// Rollback
			err = u.repository.Restore(ctxE, id)
			_ = u.log.Log("method", "author.interactor.delete", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.delete", "msg", domain.AuthorRemoved+" event published")
		}

		errC <- err
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
	}

	return nil
}

// Resiliency

// Restore recover an author from the store
func (u *Author) Restore(ctx context.Context, id string) error {
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err := u.repository.Restore(ctxR, id)
	if err != nil {
		return err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// Propagate side-effects
	errC := make(chan error)
	defer close(errC)
	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()
		if err = u.event.Restored(ctxE, id); err != nil {
			_ = u.log.Log("method", "author.interactor.restore", "err", err.Error())

			// Rollback
			err = u.repository.Remove(ctxE, id)
			_ = u.log.Log("method", "author.interactor.restore", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.restore", "msg", domain.AuthorRestored+" event published")
		}

		errC <- err
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
	}

	return err
}

// HardDelete Remove an author from the store permanently
func (u *Author) HardDelete(ctx context.Context, id string) error {
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	// Get backup
	// Using repository directly to avoid non-organic total_views increment
	author, err := u.repository.FetchByID(ctxR, id, true)
	authorBackup := author
	if err != nil {
		return err
	}

	err = u.repository.HardRemove(ctxR, id)
	if err != nil {
		return err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// Propagate side-effects
	errC := make(chan error)
	defer close(errC)
	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()
		if err = u.event.HardRemoved(ctxE, id); err != nil {
			_ = u.log.Log("method", "author.interactor.hard_delete", "err", err.Error())

			// Rollback
			err = u.repository.SaveRaw(ctxE, *authorBackup)
			_ = u.log.Log("method", "author.interactor.restore", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.hard_delete", "msg", domain.AuthorHardRemoved+" event published")
		}

		errC <- err
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
	}

	return err
}

package interactor

import (
	"context"
	"encoding/json"
	"errors"
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
	event      domain.IAuthorEventBus
}

// NewAuthorUseCase Create a new author interact
func NewAuthorUseCase(logger log.Logger, repository domain.IAuthorRepository, bus domain.IAuthorEventBus) *AuthorUseCase {
	return &AuthorUseCase{logger, repository, bus}
}

// Create Store a new entity
func (u *AuthorUseCase) Create(ctx context.Context, aggregate *domain.AuthorAggregate) (*domain.Author, error) {
	author := domain.NewAuthor(aggregate.FirstName, aggregate.LastName, aggregate.DisplayName, aggregate.OwnershipType, aggregate.OwnerID)
	err := author.IsValid()
	if err != nil {
		return nil, err
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err = u.repository.Save(ctxR, author)
	if err != nil {
		return nil, err
	}

	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// Transaction/interaction event, required owner/user validation, use concurrent-safe routine
	errChan := make(chan error)
	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()

		err = u.event.StartCreate(ctxE, author)
		if err != nil {
			_ = u.log.Log("method", "author.interactor.create", "err", err.Error())

			// Rollback
			err = u.repository.HardRemove(ctxE, author.ExternalID)
			if err != nil {
				_ = u.log.Log("method", "author.interactor.create", "err", err.Error())
			}

			_ = u.log.Log("method", "author.interactor.create", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.create", "msg", "AUTHOR_PENDING integration event published")
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
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	params := core.NewPaginationParams(pageToken, pageSize)
	params.Size++
	output, err = u.repository.Fetch(ctxR, params, filterParams)

	nextToken = ""
	if len(output) >= params.Size {
		nextToken = output[len(output)-1].ExternalID
		output = output[0 : len(output)-1]
	}
	return
}

// Get Obtain one author
func (u *AuthorUseCase) Get(ctx context.Context, id string) (*domain.Author, error) {
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	author, err := u.repository.FetchByID(ctxR, id, false)
	if err != nil {
		return nil, err
	}

	if author != nil {
		author.TotalViews++
		// Using repo directly to avoid unused fields on use case's update
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
		aggregate.RootAggregate.OwnershipType == "" && aggregate.RootAggregate.OwnerID == "" {
		return nil, exception.EmptyBody
	}

	// Get previous version
	// Using repository directly to avoid non-organic total_views increment
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	author, err := u.repository.FetchByID(ctxR, aggregate.ID, false)
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
	// If new owner id was given, then set author state to pending to start proper
	// transaction
	if aggregate.RootAggregate.OwnerID != "" {
		author.OwnerID = aggregate.RootAggregate.OwnerID
		author.Status = string(domain.StatePending)
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

	author.UpdateTime = time.Now()

	err = author.IsValid()
	if err != nil {
		return nil, err
	}

	err = u.repository.Replace(ctxR, author)
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
		if author.Status == string(domain.StatePending) {
			err = u.event.StartUpdate(ctxE, author, authorBackup)
			if err == nil {
				eventStr = "AUTHOR_UPDATE_PENDING event published"
			}
		} else {
			err = u.event.Updated(ctxE, author)
			if err == nil {
				eventStr = "AUTHOR_UPDATED event published"
			}
		}

		if err != nil {
			_ = u.log.Log("method", "author.interactor.update", "err", err.Error())

			// Rollback
			err = u.repository.Replace(ctxE, authorBackup)
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
func (u *AuthorUseCase) Delete(ctx context.Context, id string) error {
	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err := u.repository.Remove(ctxR, id)
	if err != nil {
		return err
	}
	// Domain Event nomenclature -> APP_NAME.SERVICE.ACTION
	// TODO: Recommended to move domain events and misc to infrastructure layer, use SQL transactions to handle operations atomically
	// Propagate side-effects
	errC := make(chan error)
	defer close(errC)
	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()
		if err = u.event.Deleted(ctxE, id, false); err != nil {
			_ = u.log.Log("method", "author.interactor.delete", "err", err.Error())

			// Rollback
			err = u.repository.Restore(ctxE, id)
			_ = u.log.Log("method", "author.interactor.delete", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.delete", "msg", "AUTHOR_DELETED event published")
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
func (u *AuthorUseCase) Restore(ctx context.Context, id string) error {
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
			_ = u.log.Log("method", "author.interactor.restore", "msg", "AUTHOR_RESTORED event published")
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
func (u *AuthorUseCase) HardDelete(ctx context.Context, id string) error {
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
		if err = u.event.Deleted(ctxE, id, true); err != nil {
			_ = u.log.Log("method", "author.interactor.hard_delete", "err", err.Error())

			// Rollback
			err = u.repository.SaveRaw(ctxE, authorBackup)
			_ = u.log.Log("method", "author.interactor.restore", "msg", "could not send event, rolled back")
		} else {
			_ = u.log.Log("method", "author.interactor.hard_delete", "msg", "AUTHOR_PERMANENTLY_DELETED event published")
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

// SAGA Transactions

// Done Change author's state to done, mostly for SAGA transactions
func (u *AuthorUseCase) Done(ctx context.Context, id, op string) error {
	if op != domain.AuthorCreated && op != domain.AuthorUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.AuthorCreated+" or "+domain.AuthorUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err := u.repository.ChangeState(ctxR, id, string(domain.StateDone))
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

		// Get author to properly propagate side-effects with respective payload
		// Using repo directly to avoid non-organic views
		var eventStr string
		author, err := u.repository.FetchByID(ctxE, id, false)
		if err == nil {
			if op == domain.AuthorCreated {
				err = u.event.Created(ctxE, author)
				eventStr = "AUTHOR_CREATED event published"
			} else if op == domain.AuthorUpdated {
				err = u.event.Updated(ctxE, author)
				eventStr = "AUTHOR_UPDATED event published"
			}
		}

		if err != nil {
			_ = u.log.Log("method", "author.interactor.done", "err", err.Error())

			// Rollback
			err = u.repository.ChangeState(ctxE, id, string(domain.StatePending))
			_ = u.log.Log("method", "author.interactor.done", "msg", "could not send event, rolled back")

		} else {
			_ = u.log.Log("method", "author.interactor.done", "msg", eventStr)
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

// Failed Restore or hard delete author for rollback, mostly for SAGA transactions
func (u *AuthorUseCase) Failed(ctx context.Context, id, op, backup string) error {
	if op != domain.AuthorCreated && op != domain.AuthorUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.AuthorCreated+" or "+domain.AuthorUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	var err error
	if op == domain.AuthorCreated {
		err = u.repository.HardRemove(ctxR, id)
	} else if op == domain.AuthorUpdated {
		authorBackup := new(domain.Author)
		err = json.Unmarshal([]byte(backup), authorBackup)
		if err != nil {
			return err
		}

		err = u.repository.Replace(ctxR, authorBackup)
	}

	// Avoid not found errors to send acknowledgement to broker
	if err != nil && errors.Unwrap(err) != exception.EntityNotFound {
		_ = u.log.Log("method", "author.interactor.failed", "err", err.Error())
		return err
	}

	_ = u.log.Log("method", "author.interactor.failed", "msg", fmt.Sprintf("author %s rolled back", id),
		"operation", op)

	return nil
}

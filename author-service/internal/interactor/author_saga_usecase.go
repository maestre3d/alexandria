package interactor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
)

type AuthorSAGA struct {
	logger     log.Logger
	repository domain.AuthorRepository
	eventSAGA  domain.AuthorSAGAEventBus
	eventBus   domain.AuthorEventBus
}

func NewAuthorSAGA(logger log.Logger, repo domain.AuthorRepository, event domain.AuthorSAGAEventBus, eventBus domain.AuthorEventBus) *AuthorSAGA {
	return &AuthorSAGA{
		logger:     logger,
		repository: repo,
		eventSAGA:  event,
		eventBus:   eventBus,
	}
}

// Choreography-SAGA actions

// Verifier implementation

func (u *AuthorSAGA) Verify(ctx context.Context, authorsJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// owners contains just an array with Authors id's string
	var authors []string
	err := json.Unmarshal(authorsJSON, &authors)
	if err != nil {
		_ = u.logger.Log("method", "author.interactor.saga.verify", "err", err.Error())

		// Rollback if format error
		err = u.eventSAGA.Failed(ctxR)
		if err != nil {
			// Error during publishing
			_ = u.logger.Log("method", "author.interactor.saga.verify", "err", err.Error())
			return err
		}

		_ = u.logger.Log("method", "author.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	for _, id := range authors {
		_, err := u.repository.FetchByID(ctxR, id, false)
		if err != nil {
			_ = u.logger.Log("method", "author.interactor.saga.verify", "err", err.Error())

			err = u.eventSAGA.Failed(ctxR)
			if err != nil {
				_ = u.logger.Log("method", "author.interactor.saga.verify", "err", err.Error())
				return err
			}

			_ = u.logger.Log("method", "author.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
			return nil
		}
	}

	// All Authors have been verified
	err = u.eventSAGA.Verified(ctxR)
	if err != nil {
		_ = u.logger.Log("method", "author.interactor.saga.verify", "err", err.Error())
		return err
	}

	_ = u.logger.Log("method", "author.interactor.saga.verify", "msg", domain.OwnerVerified+" integration event published")
	return nil
}

// Finishing actions

func (u *AuthorSAGA) Done(ctx context.Context, rootID, operation string) error {
	if operation != domain.AuthorCreated && operation != domain.AuthorUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.AuthorCreated+" or "+domain.AuthorUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err := u.repository.ChangeState(ctxR, rootID, domain.StatusDone)
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
		author, err := u.repository.FetchByID(ctxE, rootID, false)
		if err != nil {
			errC <- err
			return
		}

		event := domain.AuthorCreated
		if operation == domain.AuthorCreated {
			err = u.eventSAGA.Created(ctxE, *author)
		} else if operation == domain.AuthorUpdated {
			err = u.eventBus.Updated(ctxE, *author)
			event = domain.AuthorUpdated
		}
		if err != nil {
			_ = u.logger.Log("method", "author.interactor.saga.done", "err", err.Error())

			// Rollback
			err = u.repository.ChangeState(ctxE, rootID, domain.StatusPending)
			if err != nil {
				// Failed to rollback
				_ = u.logger.Log("method", "author.interactor.saga.done", "err", err.Error())
				errC <- err
				return
			}
			_ = u.logger.Log("method", "author.interactor.saga.done", "msg", "could not send event, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "author.interactor.saga.done", "msg", event+" event published")

		errC <- nil
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
func (u *AuthorSAGA) Failed(ctx context.Context, rootID, operation, backup string) error {
	if operation != domain.AuthorCreated && operation != domain.AuthorUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.AuthorCreated+" or "+domain.AuthorUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	var err error
	if operation == domain.AuthorCreated {
		err = u.repository.HardRemove(ctxR, rootID)
	} else if operation == domain.AuthorUpdated {
		authorBackup := new(domain.Author)
		err = json.Unmarshal([]byte(backup), authorBackup)
		if err != nil {
			return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"backup", "backup entity"))
		}

		err = u.repository.Replace(ctxR, *authorBackup)
	}

	// Avoid not found errors to send acknowledgement to broker
	if err != nil && !errors.Is(err, exception.EntityNotFound) {
		_ = u.logger.Log("method", "author.interactor.failed", "err", err.Error())
		return err
	}

	_ = u.logger.Log("method", "author.interactor.failed", "msg", fmt.Sprintf("author %s rolled back", rootID),
		"operation", operation)

	return nil
}

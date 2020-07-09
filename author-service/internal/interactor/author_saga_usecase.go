package interactor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"strings"
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

func (u *AuthorSAGA) UpdatePicture(ctx context.Context, rootID string, urlJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	author, err := u.repository.FetchByID(ctxR, rootID, false)
	if code := httputil.ErrorToCode(err); err != nil && code != 500 {
		// Rollback if user (e.g. HTTP 404) error
		errE := u.eventSAGA.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "author.interactor.saga.update_picture", "msg", domain.BlobFailed+" integration event published")
		return err
	}

	urls := []string{}
	err = json.Unmarshal(urlJSON, &urls)
	if err != nil {
		// Rollback if user (e.g. HTTP 404/409/400) error
		err = u.eventSAGA.BlobFailed(ctxR, err.Error())
		if err != nil {
			// Error during publishing
			return err
		}

		_ = u.logger.Log("method", "author.interactor.saga.update_picture", "msg", domain.BlobFailed+" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"url", "[]string"))
	}
	author.Picture = &urls[0]

	err = u.repository.Replace(ctxR, *author)
	if code := httputil.ErrorToCode(err); err != nil && code != 500 {
		// Rollback if user (e.g. HTTP 404/409/400) error
		errE := u.eventSAGA.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "author.interactor.saga.update_picture", "msg", domain.BlobFailed+" integration event published")
		return err
	}

	return nil
}

func (u *AuthorSAGA) RemovePicture(ctx context.Context, rootID []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	var rootPool []string
	err := json.Unmarshal(rootID, &rootPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"url", "[]string"))
	}

	author, err := u.repository.FetchByID(ctxR, rootPool[0], false)
	if err != nil {
		return err
	}
	author.Picture = nil

	return u.repository.Replace(ctxR, *author)
}

// Verifier implementation

func (u *AuthorSAGA) Verify(ctx context.Context, service string, authorsJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// owners contains just an array with Authors id's string
	var authors []string
	err := json.Unmarshal(authorsJSON, &authors)
	if err != nil {
		// Rollback if format error
		err = u.eventSAGA.Failed(ctxR, service, err.Error())
		if err != nil {
			// Error during publishing
			return err
		}

		_ = u.logger.Log("method", "author.interactor.saga.verify", "msg", strings.ToUpper(service)+"_"+domain.AuthorFailed+
			" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	for _, id := range authors {
		_, err := u.repository.FetchByID(ctxR, id, false)
		if code := httputil.ErrorToCode(err); err != nil && code != 500 {
			// Rollback if user (e.g. HTTP 404) error
			errE := u.eventSAGA.Failed(ctxR, service, err.Error())
			if errE != nil {
				// Error during publishing
				return errE
			}

			_ = u.logger.Log("method", "author.interactor.saga.verify", "msg", strings.ToUpper(service)+"_"+domain.AuthorFailed+
				" integration event published")
			return err
		}
	}

	// All Authors have been verified
	err = u.eventSAGA.Verified(ctxR, service)
	if err != nil {
		return err
	}

	_ = u.logger.Log("method", "author.interactor.saga.verify", "msg", strings.ToUpper(service)+"_"+domain.AuthorVerified+
		" integration event published")
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
			// Rollback
			err = u.repository.ChangeState(ctxE, rootID, domain.StatusPending)
			if err != nil {
				// Failed to rollback
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
func (u *AuthorSAGA) Failed(ctx context.Context, rootID, operation, snapshot string) error {
	if operation != domain.AuthorCreated && operation != domain.AuthorUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.AuthorCreated+" or "+domain.AuthorUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	// Perform preferred local rollback
	var err error
	if operation == domain.AuthorCreated {
		err = u.repository.HardRemove(ctxR, rootID)
	} else if operation == domain.AuthorUpdated {
		authorSnapshot := new(domain.Author)
		err = json.Unmarshal([]byte(snapshot), authorSnapshot)
		if err != nil {
			return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"snapshot", "snapshot entity"))
		}

		err = u.repository.Replace(ctxR, *authorSnapshot)
	}

	// Avoid not found errors to send acknowledgement to broker
	if err != nil && !errors.Is(err, exception.EntityNotFound) {
		return err
	}

	_ = u.logger.Log("method", "author.interactor.failed", "msg", fmt.Sprintf("author %s rolled back", rootID),
		"operation", operation)

	return nil
}

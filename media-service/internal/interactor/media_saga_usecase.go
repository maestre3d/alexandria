package interactor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
)

type MediaSAGA struct {
	repository domain.MediaRepository
	eventBus   domain.MediaEvent
	eventSAGA  domain.MediaEventSAGA
	logger     log.Logger
}

func NewMediaSAGA(repo domain.MediaRepository, ev domain.MediaEvent, es domain.MediaEventSAGA, logger log.Logger) *MediaSAGA {
	return &MediaSAGA{
		repository: repo,
		eventBus:   ev,
		eventSAGA:  es,
		logger:     logger,
	}
}

// Choreography-SAGA actions

func (u *MediaSAGA) VerifyAuthor(ctx context.Context, rootID string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	media, err := u.repository.FetchByID(ctxR, rootID, false)
	if err != nil {
		_ = u.logger.Log("method", "media.interactor.saga.verify_author", "err", err.Error())
		return err
	}

	authorPool := make([]string, 0)
	authorPool = append(authorPool, media.AuthorID)
	err = u.eventSAGA.VerifyAuthor(ctxR, authorPool)
	if err != nil {
		_ = u.logger.Log("method", "media.interactor.saga.verify_author", "err", err.Error())
	}

	_ = u.logger.Log("method", "media.interactor.saga.verify_author", "msg", domain.AuthorVerify+" event published")
	return nil
}

func (u *MediaSAGA) UpdateStatic(ctx context.Context, rootID string, urlJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	media, err := u.repository.FetchByID(ctxR, rootID, false)
	if code := httputil.ErrorToCode(err); err != nil && code != 500 {
		// Rollback if user (e.g. HTTP 404) error
		errE := u.eventSAGA.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "media.interactor.saga.update_static", "msg", domain.BlobFailed+" integration event published")
		return err
	}

	var urlPool []string
	err = json.Unmarshal(urlJSON, &urlPool)
	if err != nil {
		// Rollback if user (e.g. HTTP 404) error
		errE := u.eventSAGA.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "media.interactor.saga.update_static", "msg", domain.BlobFailed+" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"url", "[]string"))
	}

	media.ContentURL = &urlPool[0]

	err = u.repository.Replace(ctxR, *media)
	if code := httputil.ErrorToCode(err); err != nil && code != 500 {
		// Rollback if user (e.g. HTTP 404) error
		errE := u.eventSAGA.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "media.interactor.saga.update_static", "msg", domain.BlobFailed+" integration event published")
		return err
	}

	return nil
}

func (u *MediaSAGA) RemoveStatic(ctx context.Context, rootID []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	var rootPool []string
	err := json.Unmarshal(rootID, &rootPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"url", "[]string"))
	}

	// Domain event (side-effect) does not contain a rollback operation
	media, err := u.repository.FetchByID(ctxR, rootPool[0], false)
	if err != nil {
		return err
	}
	media.ContentURL = nil

	return u.repository.Replace(ctxR, *media)
}

// Verifier implementation

// Finishing actions

func (u *MediaSAGA) Done(ctx context.Context, rootID, operation string) error {
	if operation != domain.MediaCreated && operation != domain.MediaUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.MediaCreated+" or "+domain.MediaUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err := u.repository.ChangeState(ctxR, rootID, domain.StatusDone)
	if err != nil {
		return err
	}

	// Propagate side-effects
	errC := make(chan error)
	defer close(errC)

	go func() {
		ctxE, cl := context.WithCancel(ctx)
		defer cl()

		// Get author to properly propagate side-effects with respective payload
		// Using repo directly to avoid non-organic views
		media, err := u.repository.FetchByID(ctxE, rootID, false)
		if err != nil {
			errC <- err
			return
		}

		event := domain.MediaCreated
		if operation == domain.MediaCreated {
			err = u.eventSAGA.Created(ctxE, *media)
		} else if operation == domain.MediaUpdated {
			err = u.eventBus.Updated(ctxE, *media)
			event = domain.MediaUpdated
		}
		if err != nil {
			// Rollback
			err = u.repository.ChangeState(ctxE, rootID, domain.StatusPending)
			if err != nil {
				// Failed to rollback
				errC <- err
				return
			}
			_ = u.logger.Log("method", "media.interactor.saga.done", "msg", "could not send event, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.saga.done", "msg", event+" event published")
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

func (u *MediaSAGA) Failed(ctx context.Context, rootID, operation, snapshot string) (err error) {
	if operation != domain.MediaCreated && operation != domain.MediaUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.MediaCreated+" or "+domain.MediaUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	if operation == domain.MediaCreated {
		err = u.repository.HardRemove(ctxR, rootID)
	} else if operation == domain.MediaUpdated {
		mediaSnapshot := new(domain.Media)
		err = json.Unmarshal([]byte(snapshot), mediaSnapshot)
		if err != nil {
			return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
				"snapshot", "media entity"))
		}

		err = u.repository.Replace(ctxR, *mediaSnapshot)
	}

	// Avoid not found errors to send acknowledgement to broker
	if err != nil && !errors.Is(err, exception.EntityNotFound) {
		return err
	}

	_ = u.logger.Log("method", "media.interactor.saga.failed", "msg", fmt.Sprintf("media %s rolled back", rootID),
		"operation", operation)
	return err
}

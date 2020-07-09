package interactor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
)

type UserSAGA struct {
	logger     log.Logger
	repository domain.UserRepository
	event      domain.UserEventSAGA
}

func NewUserSAGA(logger log.Logger, repo domain.UserRepository, event domain.UserEventSAGA) *UserSAGA {
	return &UserSAGA{
		logger:     logger,
		repository: repo,
		event:      event,
	}
}

func (u *UserSAGA) UpdatePicture(ctx context.Context, id string, urlJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	var urls []string
	err := json.Unmarshal(urlJSON, &urls)
	if err != nil {
		// Rollback if user (e.g. HTTP 404) error
		errE := u.event.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "identity.interactor.saga.update_picture", "msg", domain.BlobFailed+" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"url", "[]string"))
	}

	user, err := u.repository.FetchByID(ctxR, id)
	if code := httputil.ErrorToCode(err); err != nil && code != 500 {
		// Rollback if user (e.g. HTTP 404) error
		errE := u.event.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "identity.interactor.saga.update_picture", "msg", domain.BlobFailed+" integration event published")
		return err
	}

	err = u.repository.ReplacePicture(ctxR, user.Username, urls[0])
	if code := httputil.ErrorToCode(err); err != nil && code != 500 {
		// Rollback if user (e.g. HTTP 404) error
		errE := u.event.BlobFailed(ctxR, err.Error())
		if errE != nil {
			// Error during publishing
			return errE
		}

		_ = u.logger.Log("method", "identity.interactor.saga.update_picture", "msg", domain.BlobFailed+" integration event published")
		return err
	}

	return nil
}

func (u *UserSAGA) RemovePicture(ctx context.Context, rootJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// Domain event, no remote rollbacks
	var rooIDPool []string
	err := json.Unmarshal(rootJSON, &rooIDPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"url", "[]string"))
	}

	user, err := u.repository.FetchByID(ctxR, rooIDPool[0])
	if err != nil {
		return err
	}

	err = u.repository.ReplacePicture(ctxR, user.Username, "")
	if err != nil {
		return err
	}

	return nil
}

// Verifier implementation

func (u *UserSAGA) Verify(ctx context.Context, service string, usersJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// owners contains just an array with users id's string
	var owners []string
	err := json.Unmarshal(usersJSON, &owners)
	if err != nil {
		// Rollback if format error
		err = u.event.Failed(ctxR, service, err.Error())
		if err != nil {
			// Error during publishing
			return err
		}

		_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	for _, id := range owners {
		_, err := u.repository.FetchByID(ctxR, id)
		if code := httputil.ErrorToCode(err); err != nil && code != 500 {
			// Rollback if user (e.g. HTTP 404/409/400) error
			errE := u.event.Failed(ctxR, service, err.Error())
			if errE != nil {
				// Error during publishing
				return errE
			}

			_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
			return err
		}
	}

	// All users have been verified
	err = u.event.Verified(ctxR, service)
	if err != nil {
		return err
	}

	_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerVerified+" integration event published")
	return nil
}

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

	// If parsing error, throw rollback event
	if err != nil {
		err = u.event.BlobFailed(ctxR, err.Error())
		if err != nil {
			return err
		}
		_ = u.logger.Log("msg", "event "+domain.BlobFailed+" produced")
	}

	user, err := u.repository.FetchByID(ctxR, id)
	// If user-provoked error, then rollback
	defer func() {
		if err != nil {
			if code := httputil.ErrorToCode(err); code != 500 {
				err = u.event.BlobFailed(ctxR, err.Error())
				if err == nil {
					_ = u.logger.Log("msg", "event "+domain.BlobFailed+" produced")
				}
			}
		}
	}()
	if err != nil {
		return err
	}

	err = u.repository.ReplacePicture(ctxR, user.Username, urls[0])
	if err != nil {
		return err
	}

	return nil
}

func (u *UserSAGA) RemovePicture(ctx context.Context, rootJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	var rooIDPool []string
	err := json.Unmarshal(rootJSON, &rooIDPool)
	if err != nil {
		return err
	}

	user, err := u.repository.FetchByID(ctxR, rooIDPool[0])
	if err != nil {
		return err
	}

	err = u.repository.ReplacePicture(ctxR, user.Username, "")
	if err != nil {
		errE := u.event.BlobFailed(ctxR, err.Error())
		if errE != nil {
			return errE
		}
		_ = u.logger.Log("msg", "event "+domain.BlobFailed+" produced")
		return err
	}

	return nil
}

// Verifier implementation

func (u *UserSAGA) Verify(ctx context.Context, usersJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// owners contains just an array with users id's string
	var owners []string
	err := json.Unmarshal(usersJSON, &owners)
	if err != nil {
		_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())

		// Rollback if format error
		err = u.event.Failed(ctxR, err.Error())
		if err != nil {
			// Error during publishing
			_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())
			return err
		}

		_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	for _, id := range owners {
		_, err := u.repository.FetchByID(ctxR, id)
		if err != nil {
			_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())

			err = u.event.Failed(ctxR, err.Error())
			if err != nil {
				_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())
				return err
			}

			_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
			return nil
		}
	}

	// All users have been verified
	err = u.event.Verified(ctxR)
	if err != nil {
		_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())
		return err
	}

	_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerVerified+" integration event published")
	return nil
}

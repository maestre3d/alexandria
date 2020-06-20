package interactor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"strings"
	"time"
)

type MediaUseCase struct {
	logger     log.Logger
	repository domain.MediaRepository
	event      domain.MediaEvent
}

func NewMediaUseCase(logger log.Logger, repo domain.MediaRepository, event domain.MediaEvent) *MediaUseCase {
	return &MediaUseCase{
		logger:     logger,
		repository: repo,
		event:      event,
	}
}

func (u *MediaUseCase) Create(ctx context.Context, ag *domain.MediaAggregate) (*domain.Media, error) {
	media, err := domain.NewMedia(ag)
	if err != nil {
		return nil, err
	}
	err = media.IsValid()
	if err != nil {
		return nil, err
	}

	ctxR, _ := context.WithCancel(ctx)
	err = u.repository.Save(ctxR, *media)
	if err != nil {
		return nil, err
	}

	// Start SAGA transaction
	errC := make(chan error)
	ctxE, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err = u.event.StartCreate(ctxE, *media)
		if err != nil {
			// Event failed to be sent
			_ = u.logger.Log("method", "media.interactor.create", "err", err.Error())
			// Rollback
			err = u.repository.HardRemove(ctxE, media.ExternalID)
			if err != nil {
				// Failed to rollback
				_ = u.logger.Log("method", "media.interactor.create", "err", err.Error())
				errC <- err
				return
			}

			_ = u.logger.Log("method", "media.interactor.create", "msg", domain.OwnerVerify+" event sending failed, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.create", "msg", domain.OwnerVerify+" integration event published")
		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return nil, err
		}
		break
	}

	return media, nil
}

func (u *MediaUseCase) Get(ctx context.Context, id string) (*domain.Media, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	media, err := u.repository.FetchByID(ctxR, id, false)
	if err != nil {
		return nil, err
	}

	// Update total views organically
	if media != nil {
		media.TotalViews++
		err = u.repository.Replace(ctxR, *media)
		if err != nil {
			_ = u.logger.Log("method", "author.interactor.get", "msg", fmt.Sprintf("could not update total_views for media %s, error: %s",
				media.ExternalID, err.Error()))
		}
	}

	return media, nil
}

func (u *MediaUseCase) List(ctx context.Context, pageToken, pageSize string, filter core.FilterParams) ([]*domain.Media, string, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	params := core.NewPaginationParams(pageToken, pageSize)
	params.Size++
	medias, err := u.repository.Fetch(ctxR, *params, filter)
	if err != nil {
		return nil, "", err
	}

	nextPage := ""
	if len(medias) >= params.Size {
		nextPage = medias[len(medias)-1].ExternalID
		medias = medias[0 : len(medias)-1]
	}

	return medias, nextPage, nil
}

func (u *MediaUseCase) Update(ctx context.Context, ag *domain.MediaUpdateAggregate) (*domain.Media, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	media, err := u.repository.FetchByID(ctxR, ag.ID, false)
	// Store backup for event rollbacks
	mediaBackup := media
	if err != nil {
		return nil, err
	}

	// Update dynamically
	// TODO: Try switch statement
	if ag.Root.Title != "" {
		media.Title = ag.Root.Title
	}
	if ag.Root.DisplayName != "" {
		media.DisplayName = ag.Root.DisplayName
	}
	if ag.Root.Description != "" {
		media.Description = ag.Root.Description
	}
	if ag.Root.LanguageCode != "" {
		media.LanguageCode = strings.ToLower(ag.Root.LanguageCode)
	}
	if ag.Root.MediaType != "" {
		media.MediaType = domain.ParseMediaType(ag.Root.MediaType)
	}
	if ag.Root.PublishDate != "" {
		date, err := domain.ParseDate(ag.Root.PublishDate)
		if err != nil {
			return nil, err
		}
		media.PublishDate = date
	}
	if ag.Root.PublisherID != "" {
		media.PublisherID = ag.Root.PublisherID
		// Must execute transaction for user validation
		media.Status = domain.StatusPending
	}
	if ag.Root.AuthorID != "" {
		media.AuthorID = ag.Root.AuthorID
		// Must execute transaction for author validation
		media.Status = domain.StatusPending
	}
	if ag.URL != "" {
		media.ContentURL = &ag.URL
	}
	media.UpdateTime = time.Now()

	err = media.IsValid()
	if err != nil {
		return nil, err
	}

	// Send side-effects/transaction events
	err = u.repository.Replace(ctxR, *media)
	if err != nil {
		return nil, err
	}

	errC := make(chan error)
	ctxE, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		event := domain.OwnerVerify
		if media.Status == domain.StatusPending {
			err = u.event.StartUpdate(ctxE, *media, *mediaBackup)
		} else {
			event = domain.MediaUpdated
			err = u.event.Updated(ctxE, *media)
		}

		if err != nil {
			// Event failed to be sent
			_ = u.logger.Log("method", "media.interactor.update", "err", err.Error())
			// Rollback
			err = u.repository.Replace(ctxE, *mediaBackup)
			if err != nil {
				// Failed to rollback
				_ = u.logger.Log("method", "media.interactor.update", "err", err.Error())
				errC <- err
				return
			}

			_ = u.logger.Log("method", "media.interactor.update", "msg", event+" event sending failed, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.update", "msg", event+" integration event published")
		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return nil, err
		}
		break
	}

	return media, nil
}

func (u *MediaUseCase) Delete(ctx context.Context, id string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	err := u.repository.Remove(ctxR, id)
	if err != nil {
		return err
	}

	// Send side-effects/domain event
	errC := make(chan error)
	ctxE, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err = u.event.Removed(ctxE, id)
		if err != nil {
			// Event failed to be sent
			_ = u.logger.Log("method", "media.interactor.delete", "err", err.Error())
			// Rollback
			err = u.repository.Restore(ctxE, id)
			if err != nil {
				// Failed to rollback
				_ = u.logger.Log("method", "media.interactor.delete", "err", err.Error())
				errC <- err
				return
			}

			_ = u.logger.Log("method", "media.interactor.delete", "msg", domain.MediaDeleted+" event sending failed, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.delete", "msg", domain.MediaDeleted+" integration event published")
		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (u *MediaUseCase) Restore(ctx context.Context, id string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	err := u.repository.Restore(ctxR, id)
	if err != nil {
		return err
	}

	// Send side-effects/domain event
	errC := make(chan error)
	ctxE, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err = u.event.Restored(ctxE, id)
		if err != nil {
			// Event failed to be sent
			_ = u.logger.Log("method", "media.interactor.restore", "err", err.Error())
			// Rollback
			err = u.repository.Remove(ctxE, id)
			if err != nil {
				// Failed to rollback
				_ = u.logger.Log("method", "media.interactor.restore", "err", err.Error())
				errC <- err
				return
			}

			_ = u.logger.Log("method", "media.interactor.restore", "msg", domain.MediaRestored+" event sending failed, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.restore", "msg", domain.MediaRestored+" integration event published")
		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (u *MediaUseCase) HardDelete(ctx context.Context, id string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	media, err := u.repository.FetchByID(ctxR, id, true)
	if err != nil {
		return err
	}

	err = u.repository.HardRemove(ctxR, id)
	if err != nil {
		return err
	}

	// Send side-effects/domain event
	errC := make(chan error)
	ctxE, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err = u.event.HardRemoved(ctxE, id)
		if err != nil {
			// Event failed to be sent
			_ = u.logger.Log("method", "media.interactor.hard_delete", "err", err.Error())
			// Rollback
			err = u.repository.SaveRaw(ctxE, *media)
			if err != nil {
				// Failed to rollback
				_ = u.logger.Log("method", "media.interactor.hard_delete", "err", err.Error())
				errC <- err
				return
			}

			_ = u.logger.Log("method", "media.interactor.hard_delete", "msg", domain.MediaHardDeleted+" event sending failed, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.hard_delete", "msg", domain.MediaHardDeleted+" integration event published")
		errC <- nil
	}()

	select {
	case err = <-errC:
		if err != nil {
			return err
		}
		break
	}

	return nil
}

// SAGA Transactions

func (u *MediaUseCase) Done(ctx context.Context, id, op string) error {
	if op != domain.MediaCreated && op != domain.MediaUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.MediaCreated+" or "+domain.MediaUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	err := u.repository.ChangeState(ctxR, id, domain.StatusDone)
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
		media, err := u.repository.FetchByID(ctxE, id, false)
		if err != nil {
			_ = u.logger.Log("method", "media.interactor.done", "err", err.Error())

			// Rollback
			err = u.repository.ChangeState(ctxE, id, domain.StatusPending)
			if err != nil {
				// Failed to rollback
				_ = u.logger.Log("method", "media.interactor.done", "err", err.Error())
				errC <- err
				return
			}
			_ = u.logger.Log("method", "media.interactor.done", "msg", "could not send event, rolled back")
			errC <- err
			return
		}

		event := domain.MediaCreated
		if op == domain.MediaCreated {
			err = u.event.Created(ctxE, *media)
		} else if op == domain.MediaUpdated {
			err = u.event.Updated(ctxE, *media)
			event = domain.MediaUpdated
		}

		_ = u.logger.Log("method", "media.interactor.done", "msg", event+" event published")

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

func (u *MediaUseCase) Failed(ctx context.Context, id, op, backup string) error {
	if op != domain.MediaCreated && op != domain.MediaUpdated {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"operation", domain.MediaCreated+" or "+domain.MediaUpdated))
	}

	ctxR, cl := context.WithCancel(ctx)
	defer cl()

	var err error
	if op == domain.MediaCreated {
		err = u.repository.HardRemove(ctxR, id)
	} else if op == domain.MediaUpdated {
		mediaBackup := new(domain.Media)
		err = json.Unmarshal([]byte(backup), mediaBackup)
		if err != nil {
			return err
		}

		err = u.repository.Replace(ctxR, *mediaBackup)
	}

	// Avoid not found errors to send acknowledgement to broker
	if err != nil && errors.Unwrap(err) != exception.EntityNotFound {
		_ = u.logger.Log("method", "media.interactor.failed", "err", err.Error())
		return err
	}

	_ = u.logger.Log("method", "media.interactor.failed", "msg", fmt.Sprintf("media %s rolled back", id),
		"operation", op)

	return nil
}

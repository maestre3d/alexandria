package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"strings"
	"time"
)

type Media struct {
	logger     log.Logger
	repository domain.MediaRepository
	event      domain.MediaEvent
}

func NewMedia(logger log.Logger, repo domain.MediaRepository, event domain.MediaEvent) *Media {
	return &Media{
		logger:     logger,
		repository: repo,
		event:      event,
	}
}

func (u *Media) Create(ctx context.Context, ag *domain.MediaAggregate) (*domain.Media, error) {
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

func (u *Media) Get(ctx context.Context, id string) (*domain.Media, error) {
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

func (u *Media) List(ctx context.Context, pageToken, pageSize string, filter core.FilterParams) ([]*domain.Media, string, error) {
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

func (u *Media) Update(ctx context.Context, ag *domain.MediaUpdateAggregate) (*domain.Media, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	media, err := u.repository.FetchByID(ctxR, ag.ID, false)
	if err != nil {
		return nil, err
	}
	// Store backup for event rollbacks
	mediaBackup := media

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

func (u *Media) Delete(ctx context.Context, id string) error {
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

			_ = u.logger.Log("method", "media.interactor.delete", "msg", domain.MediaRemoved+" event sending failed, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.delete", "msg", domain.MediaRemoved+" event published")
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

func (u *Media) Restore(ctx context.Context, id string) error {
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

		_ = u.logger.Log("method", "media.interactor.restore", "msg", domain.MediaRestored+" event published")
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

func (u *Media) HardDelete(ctx context.Context, id string) error {
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

			_ = u.logger.Log("method", "media.interactor.hard_delete", "msg", domain.MediaHardRemoved+" event sending failed, rolled back")
			errC <- err
			return
		}

		_ = u.logger.Log("method", "media.interactor.hard_delete", "msg", domain.MediaHardRemoved+" event published")
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

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

type MediaUseCase struct {
	logger     log.Logger
	repository domain.MediaRepository
}

func NewMediaUseCase(logger log.Logger, repo domain.MediaRepository) *MediaUseCase {
	return &MediaUseCase{
		logger:     logger,
		repository: repo,
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
	// TODO: Send/Start SAGA transaction event in goroutine

	return media, nil
}

func (u *MediaUseCase) Get(ctx context.Context, id string) (*domain.Media, error) {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	media, err := u.repository.FetchByID(ctxR, id, true)
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
	// mediaBackup := media
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

	// TODO: Send side-effects/transaction events
	err = u.repository.Replace(ctxR, *media)
	if err != nil {
		return nil, err
	}

	return media, nil
}

func (u *MediaUseCase) Delete(ctx context.Context, id string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// TODO: Send side-effect event
	return u.repository.Remove(ctxR, id)
}

func (u *MediaUseCase) Restore(ctx context.Context, id string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// TODO: Send side-effect event
	return u.repository.Restore(ctxR, id)
}

func (u *MediaUseCase) HardDelete(ctx context.Context, id string) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// TODO: Send side-effect event
	return u.repository.HardRemove(ctxR, id)
}

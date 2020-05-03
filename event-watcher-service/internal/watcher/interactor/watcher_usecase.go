package interactor

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/shared/domain/exception"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/shared/domain/util"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/domain"
)

type WatcherUseCase struct {
	ctx        context.Context
	logger     log.Logger
	repository domain.WatcherRepository
	validate   *validator.Validate
}

func NewWatcherUseCase(ctx context.Context, logger log.Logger, repository domain.WatcherRepository) *WatcherUseCase {
	return &WatcherUseCase{ctx, logger, repository, validator.New()}
}

// Create Generate and save a new watcher entity
func (u *WatcherUseCase) Create(serviceName, transactionID, eventType, content, importance, provider string) (*domain.WatcherEntity, error) {
	watcher := domain.NewWatcherEntity(serviceName, transactionID, eventType, content, importance, provider)

	err := u.validate.Struct(watcher)
	if err != nil {
		return nil, err
	}

	err = u.repository.Save(watcher)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

// List Obtain a watcher's entities list
func (u *WatcherUseCase) List(token, size string, filterParams util.FilterParams) ([]*domain.WatcherEntity, string, error) {
	params := util.NewPaginationParams(token, size)

	// Prepare next_token_id
	params.Size += 1
	watchers, err := u.repository.Fetch(params, filterParams)
	if err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(watchers) >= params.Size {
		nextToken = watchers[len(watchers)-1].ID
		watchers = watchers[0 : len(watchers)-1]
	}

	return watchers, nextToken, nil
}

// Get Obtain an specific watcher entity
func (u *WatcherUseCase) Get(id string) (*domain.WatcherEntity, error) {
	err := validateID(id)
	if err != nil {
		return nil, err
	}

	return u.repository.FetchByID(id)
}

// Update Modify an specific watcher entity
func (u *WatcherUseCase) Update(id, serviceName, transactionID, eventType, content, importance, provider string) (*domain.WatcherEntity, error) {
	err := validateID(id)
	if err != nil {
		return nil, err
	}

	watcher, err := u.repository.FetchByID(id)
	if err != nil {
		return nil, err
	}

	// Atomic update
	switch {
	case serviceName != "":
		watcher.ServiceName = serviceName
	case transactionID != "":
		watcher.TransactionID = &transactionID
	case eventType != "":
		watcher.EventType = eventType
	case content != "":
		watcher.Content = content
	case importance != "":
		watcher.Importance = importance
	case provider != "":
		watcher.Provider = provider
	}

	err = u.validate.Struct(watcher)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

// Delete Remove an specific watcher entity
func (u *WatcherUseCase) Delete(id string) error {
	err := validateID(id)
	if err != nil {
		return err
	}

	return u.repository.Remove(id)
}

// validateID Throws an error if ID is malformed
func validateID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidID
	}

	return err
}

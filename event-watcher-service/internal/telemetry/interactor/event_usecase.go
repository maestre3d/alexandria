package interactor

import (
	"context"
	"github.com/alexandria-oss/core"
	"github.com/go-kit/kit/log"
	"github.com/go-playground/validator/v10"
	"github.com/maestre3d/alexandria/event-watcher-service/internal/watcher/domain"
	"strings"
)

type EventUseCase struct {
	ctx        context.Context
	logger     log.Logger
	repository domain.EventRepository
	validate   *validator.Validate
}

func NewEventUseCase(ctx context.Context, logger log.Logger, repository domain.EventRepository) *EventUseCase {
	return &EventUseCase{ctx, logger, repository, validator.New()}
}

// Create Generate and save a new event entity
func (u *EventUseCase) Create(serviceName, transactionID, eventType, content, importance, provider string) (*domain.EventEntity, error) {
	event := domain.NewEventEntity(serviceName, transactionID, eventType, content, importance, provider)

	err := u.validate.Struct(event)
	if err != nil {
		return nil, err
	}

	err = u.repository.Save(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// List Obtain a event's entities list
func (u *EventUseCase) List(token, size string, filterParams core.FilterParams) ([]*domain.EventEntity, string, error) {
	params := core.NewPaginationParams(token, size)
	params.Token = token

	// Prepare next_token_id
	params.Size += 1
	events, err := u.repository.Fetch(params, filterParams)
	if err != nil {
		return nil, "", err
	}

	nextToken := ""
	if len(events) >= params.Size {
		nextToken = events[len(events)-1].ID
		events = events[0 : len(events)-1]
	}

	return events, nextToken, nil
}

// Get Obtain an specific watcher entity
func (u *EventUseCase) Get(id string) (*domain.EventEntity, error) {
	return u.repository.FetchByID(id)
}

// Update Modify an specific event entity
func (u *EventUseCase) Update(id, serviceName, transactionID, eventType, content, importance, provider string) (*domain.EventEntity, error) {
	event, err := u.Get(id)
	if err != nil {
		return nil, err
	}

	// Atomic update
	if serviceName != "" {
		event.ServiceName = strings.ToUpper(serviceName)
	}
	if eventType != "" {
		event.EventType = eventType
	}
	if content != "" {
		event.Content = content
	}
	if importance != "" {
		event.Importance = importance
	}
	if provider != "" {
		event.Provider = strings.ToUpper(provider)
	}
	if transactionID != "" {
		event.TransactionID = &transactionID
	}

	err = u.validate.Struct(event)
	if err != nil {
		return nil, err
	}

	// Assign ID from params in case of NoSQL sort key
	event.ID = id
	err = u.repository.Update(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// Delete Remove an specific event entity
func (u *EventUseCase) Delete(id string) error {
	return u.repository.Remove(id)
}

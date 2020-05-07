package interactor

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/go-playground/validator/v10"
	"github.com/maestre3d/alexandria/event-telemetry-service/internal/telemetry/domain"
	"strconv"
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
func (u *EventUseCase) Create(serviceName, eventType, priority, provider string, content []byte, isTransaction bool) (*eventbus.Event, error) {

	event := eventbus.NewEvent(serviceName, eventType, priority, provider, content, isTransaction)
	/*
		err := u.validate.Struct(event)
		if err != nil {
			return nil, err
		}*/

	err := u.repository.Save(event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// CreateRaw Store a simple event entity with raw params
func (u *EventUseCase) CreateRaw(event *eventbus.Event) error {
	return u.repository.Save(event)
}

// List Obtain a event's entities list
func (u *EventUseCase) List(token, size string, filterParams core.FilterParams) ([]*eventbus.Event, string, error) {
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

// Get Obtain an specific telemetry entity
func (u *EventUseCase) Get(id string) (*eventbus.Event, error) {
	return u.repository.FetchByID(id)
}

// Update Modify an specific event entity
func (u *EventUseCase) Update(id, serviceName, transactionID, eventType, priority, provider string, content []byte) (*eventbus.Event, error) {
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
	if content != nil {
		event.Content = content
	}
	if priority != "" {
		event.Priority = priority
	}
	if provider != "" {
		event.Provider = strings.ToUpper(provider)
	}
	if transactionID != "" {
		event.TransactionID, err = strconv.ParseUint(transactionID, 10, 64)
		if err != nil {
			return nil, exception.NewErrorDescription(exception.InvalidFieldFormat,
				fmt.Sprintf(exception.InvalidFieldFormatString, "transaction_id", "uint64"))
		}
	}

	/*
		err = u.validate.Struct(event)
		if err != nil {
			return nil, err
		}*/

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

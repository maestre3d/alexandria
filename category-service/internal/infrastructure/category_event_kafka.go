package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/maestre3d/alexandria/category-service/internal/infrastructure/eventutil"
	"gocloud.dev/pubsub"
	"sync"
)

type CategoryEventKafka struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewCategoryEventKafka(cfg *config.Kernel) *CategoryEventKafka {
	return &CategoryEventKafka{
		cfg: cfg,
		mu:  new(sync.Mutex),
	}
}

func (e *CategoryEventKafka) Created(ctx context.Context, category domain.Category) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"category", "category object"))
	}

	spanJSON, err := eventutil.SpanCtxToJSON(ctx)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, categoryJSON)
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryCreated)
	if err != nil {
		return err
	}
	defer func() {
		err = p.Shutdown(ctx)
	}()

	m := &pubsub.Message{
		Body:       event.Content,
		Metadata:   eventutil.GenerateEventMetadata(*event),
		BeforeSend: nil,
	}

	return eventutil.PublishResilientEvent(ctx, eventutil.EventAggregate{
		Name:    "created",
		Prefix:  e.cfg.Service,
		Topic:   p,
		Message: m,
	})
}

func (e *CategoryEventKafka) Updated(ctx context.Context, category domain.Category) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"category", "category object"))
	}

	spanJSON, err := eventutil.SpanCtxToJSON(ctx)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, categoryJSON)
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryUpdated)
	if err != nil {
		return err
	}
	defer func() {
		err = p.Shutdown(ctx)
	}()

	m := &pubsub.Message{
		Body:       event.Content,
		Metadata:   eventutil.GenerateEventMetadata(*event),
		BeforeSend: nil,
	}

	return eventutil.PublishResilientEvent(ctx, eventutil.EventAggregate{
		Name:    "updated",
		Prefix:  e.cfg.Service,
		Topic:   p,
		Message: m,
	})
}

func (e *CategoryEventKafka) Removed(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	spanJSON, err := eventutil.SpanCtxToJSON(ctx)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryRemoved)
	if err != nil {
		return err
	}
	defer func() {
		err = p.Shutdown(ctx)
	}()

	m := &pubsub.Message{
		Body:       event.Content,
		Metadata:   eventutil.GenerateEventMetadata(*event),
		BeforeSend: nil,
	}

	return eventutil.PublishResilientEvent(ctx, eventutil.EventAggregate{
		Name:    "removed",
		Prefix:  e.cfg.Service,
		Topic:   p,
		Message: m,
	})
}

func (e *CategoryEventKafka) Restored(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	spanJSON, err := eventutil.SpanCtxToJSON(ctx)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryRestored)
	if err != nil {
		return err
	}
	defer func() {
		err = p.Shutdown(ctx)
	}()

	m := &pubsub.Message{
		Body:       event.Content,
		Metadata:   eventutil.GenerateEventMetadata(*event),
		BeforeSend: nil,
	}

	return eventutil.PublishResilientEvent(ctx, eventutil.EventAggregate{
		Name:    "restored",
		Prefix:  e.cfg.Service,
		Topic:   p,
		Message: m,
	})
}

func (e *CategoryEventKafka) HardRemoved(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	spanJSON, err := eventutil.SpanCtxToJSON(ctx)
	if err != nil {
		return err
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryHardRemoved)
	if err != nil {
		return err
	}
	defer func() {
		err = p.Shutdown(ctx)
	}()

	m := &pubsub.Message{
		Body:       event.Content,
		Metadata:   eventutil.GenerateEventMetadata(*event),
		BeforeSend: nil,
	}

	return eventutil.PublishResilientEvent(ctx, eventutil.EventAggregate{
		Name:    "hard_removed",
		Prefix:  e.cfg.Service,
		Topic:   p,
		Message: m,
	})
}

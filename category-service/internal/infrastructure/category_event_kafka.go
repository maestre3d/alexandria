package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/category-service/internal/domain"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
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

func getCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          name,
		MaxRequests:   1,
		Interval:      0,
		Timeout:       0,
		ReadyToTrip:   nil,
		OnStateChange: nil,
	})
}

func (e *CategoryEventKafka) Created(ctx context.Context, category domain.Category) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"category", "category object"))
	}

	ctxT, span := trace.StartSpan(ctx, "category_created")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryCreated), trace.StringAttribute("event.type", eventbus.EventDomain))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context object"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, categoryJSON)
	event.TracingContext = string(spanJSON)
	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryCreated)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"event_id":        event.ID,
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	_, err = getCircuitBreaker("category_created").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})
	return err
}

func (e *CategoryEventKafka) Updated(ctx context.Context, category domain.Category) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	categoryJSON, err := json.Marshal(category)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"category", "category object"))
	}

	ctxT, span := trace.StartSpan(ctx, "category_updated")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryUpdated), trace.StringAttribute("event.type", eventbus.EventDomain))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context object"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, categoryJSON)
	event.TracingContext = string(spanJSON)
	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryUpdated)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"event_id":        event.ID,
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	_, err = getCircuitBreaker("category_updated").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})
	return err
}

func (e *CategoryEventKafka) Removed(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctxT, span := trace.StartSpan(ctx, "category_removed")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryRemoved), trace.StringAttribute("event.type", eventbus.EventDomain))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context object"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)
	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryRemoved)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"event_id":        event.ID,
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	_, err = getCircuitBreaker("category_removed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})
	return err
}

func (e *CategoryEventKafka) HardRemoved(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctxT, span := trace.StartSpan(ctx, "category_hard_removed")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.CategoryHardRemoved), trace.StringAttribute("event.type", eventbus.EventDomain))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context object"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.CategoryHardRemoved)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"event_id":        event.ID,
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	_, err = getCircuitBreaker("category_hard_removed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})
	return err
}

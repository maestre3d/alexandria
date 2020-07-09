package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/google/uuid"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"sync"
	"time"
)

type MediaKafkaEvent struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewMediaKafakaEvent(cfg *config.Kernel) *MediaKafkaEvent {
	return &MediaKafkaEvent{mu: new(sync.Mutex), cfg: cfg}
}

func (e MediaKafkaEvent) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "media_kafka_" + action,
		MaxRequests: 1,
		Interval:    0,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: nil,
	}

	return gobreaker.NewCircuitBreaker(st)
}

func (e *MediaKafkaEvent) StartCreate(ctx context.Context, media domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ownerPool := make([]string, 0)
	ownerPool = append(ownerPool, media.PublisherID)

	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: start_create")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.OwnerVerify))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, ownerJSON)
	event.TracingContext = string(spanJSON)
	t := eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    media.ExternalID,
		SpanID:    span.SpanContext().SpanID.String(),
		TraceID:   span.SpanContext().TraceID.String(),
		Operation: domain.MediaCreated,
	}
	p, err := eventbus.NewKafkaProducer(ctx, domain.OwnerVerify)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id":  t.ID,
			"root_id":         t.RootID,
			"span_id":         t.SpanID,
			"trace_id":        t.TraceID,
			"operation":       t.Operation,
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("start_create").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *MediaKafkaEvent) StartUpdate(ctx context.Context, media domain.Media, snapshot domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ownerPool := make([]string, 0)
	ownerPool = append(ownerPool, media.PublisherID)
	ownerJSON, err := json.Marshal(ownerPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"snapshot", "media entity"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: start_update")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.OwnerVerify))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	t := &eventbus.Transaction{
		ID:        uuid.New().String(),
		RootID:    media.ExternalID,
		SpanID:    span.SpanContext().SpanID.String(),
		TraceID:   span.SpanContext().TraceID.String(),
		Operation: domain.MediaUpdated,
		Snapshot:  string(snapshotJSON),
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, ownerJSON)
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.OwnerVerify)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id":  t.ID,
			"root_id":         t.RootID,
			"span_id":         t.SpanID,
			"trace_id":        t.TraceID,
			"operation":       t.Operation,
			"snapshot":        t.Snapshot,
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("start_update").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *MediaKafkaEvent) Updated(ctx context.Context, media domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"media", "media entity"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: updated")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.MediaUpdated))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, mediaJSON)
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaUpdated)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("updated").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *MediaKafkaEvent) Removed(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: removed")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.MediaRemoved))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaRemoved)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("removed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *MediaKafkaEvent) Restored(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: restored")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.MediaRestored))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaRestored)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("restored").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *MediaKafkaEvent) HardRemoved(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: hard_removed")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.MediaHardRemoved))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityMid, eventbus.ProviderKafka, []byte(id))
	event.TracingContext = string(spanJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaHardRemoved)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"tracing_context": event.TracingContext,
			"service":         event.ServiceName,
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("hard_removed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

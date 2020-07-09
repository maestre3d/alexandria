package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"sync"
	"time"
)

type MediaSAGAKafkaEvent struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewMediaSAGAKafkaEvent(cfg *config.Kernel) *MediaSAGAKafkaEvent {
	return &MediaSAGAKafkaEvent{
		cfg: cfg,
		mu:  new(sync.Mutex),
	}
}

func (e MediaSAGAKafkaEvent) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "media_saga_kafka_" + action,
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

func (e *MediaSAGAKafkaEvent) VerifyAuthor(ctx context.Context, authorPool []string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Owner/User verified, publish SERVICE_OWNER_VERIFIED
	ec, err := eventbus.ExtractContext(ctx)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"event", "event context"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: verify")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.AuthorVerify))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	authorJSON, err := json.Marshal(authorPool)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"author_pool", "[]string"))
	}

	p, err := eventbus.NewKafkaProducer(ctx, domain.AuthorVerify)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	ec.Transaction.SpanID = span.SpanContext().SpanID.String()
	ec.Transaction.TraceID = span.SpanContext().TraceID.String()

	event := eventbus.NewEvent(e.cfg.Service, ec.Event.EventType, ec.Event.Priority, eventbus.ProviderKafka, authorJSON)
	event.TracingContext = string(spanJSON)
	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id":  ec.Transaction.ID,
			"root_id":         ec.Transaction.RootID,
			"span_id":         ec.Transaction.SpanID,
			"trace_id":        ec.Transaction.TraceID,
			"operation":       ec.Transaction.Operation,
			"snapshot":        ec.Transaction.Snapshot,
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

func (e *MediaSAGAKafkaEvent) Created(ctx context.Context, media domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: created")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.MediaCreated))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"media", "media entity"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, mediaJSON)
	event.TracingContext = string(spanJSON)
	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaCreated)
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
	_, err = e.defaultCircuitBreaker("created").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *MediaSAGAKafkaEvent) BlobFailed(ctx context.Context, msg string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ec, err := eventbus.ExtractContext(ctx)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"event", "event context"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "media: created")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.MediaCreated))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	p, err := eventbus.NewKafkaProducer(ctx, domain.BlobFailed)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	ec.Transaction.SpanID = span.SpanContext().SpanID.String()
	ec.Transaction.TraceID = span.SpanContext().TraceID.String()

	ev := eventbus.NewEvent(e.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, []byte(msg))
	ev.TracingContext = string(spanJSON)
	m := &pubsub.Message{
		Body: ev.Content,
		Metadata: map[string]string{
			"transaction_id":  ec.Transaction.ID,
			"root_id":         ec.Transaction.RootID,
			"span_id":         ec.Transaction.SpanID,
			"trace_id":        ec.Transaction.TraceID,
			"operation":       ec.Transaction.Operation,
			"snapshot":        ec.Transaction.Snapshot,
			"tracing_context": ev.TracingContext,
			"service":         ev.ServiceName,
			"event_id":        ev.ID,
			"event_type":      ev.EventType,
			"priority":        ev.Priority,
			"provider":        ev.Provider,
			"dispatch_time":   ev.DispatchTime,
		},
		BeforeSend: nil,
	}

	_, err = e.defaultCircuitBreaker("blob_failed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"strings"
	"sync"
	"time"
)

type AuthorSAGAKafkaEventBus struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewAuthorSAGAKafkaEventBus(cfg *config.Kernel) *AuthorSAGAKafkaEventBus {
	return &AuthorSAGAKafkaEventBus{
		cfg: cfg,
		mu:  new(sync.Mutex),
	}
}

func (e AuthorSAGAKafkaEventBus) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "author_saga_kafka_" + action,
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

func (e *AuthorSAGAKafkaEventBus) Verified(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Owner/User verified, publish SERVICE_OWNER_VERIFIED
	eC, err := eventbus.ExtractContext(ctx)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"event", "event context"))
	}

	// Avoid non-service naming, it would be impossible to respond to event
	if eC.Event.ServiceName == "" {
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "service_name"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "author: verified")
	defer span.End()

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", strings.ToUpper(eC.Event.ServiceName)+"_"+domain.AuthorVerified))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	p, err := eventbus.NewKafkaProducer(ctxT, strings.ToUpper(eC.Event.ServiceName)+"_"+domain.AuthorVerified)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctxT)

	eC.Transaction.SpanID = span.SpanContext().SpanID.String()
	eC.Transaction.TraceID = span.SpanContext().TraceID.String()

	event := eventbus.NewEvent(e.cfg.Service, eC.Event.EventType, eC.Event.Priority, eventbus.ProviderKafka, []byte(""))
	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id":  eC.Transaction.ID,
			"root_id":         eC.Transaction.RootID,
			"span_id":         eC.Transaction.SpanID,
			"trace_id":        eC.Transaction.TraceID,
			"operation":       eC.Transaction.Operation,
			"snapshot":        eC.Transaction.Snapshot,
			"tracing_context": string(spanJSON),
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
	_, err = e.defaultCircuitBreaker("author_verified").Execute(func() (interface{}, error) {
		return nil, p.Send(ctxT, m)
	})

	return err
}

func (e *AuthorSAGAKafkaEventBus) Failed(ctx context.Context, msg string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Owner/User verified, publish SERVICE_OWNER_VERIFIED
	eC, err := eventbus.ExtractContext(ctx)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"event", "event context"))
	}

	// Avoid non-service naming, it would be impossible to respond to event
	if eC.Event.ServiceName == "" {
		return exception.NewErrorDescription(exception.RequiredField, fmt.Sprintf(exception.RequiredFieldString, "service_name"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "author: failed")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", strings.ToUpper(eC.Event.ServiceName)+""+domain.AuthorFailed))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	p, err := eventbus.NewKafkaProducer(ctx, strings.ToUpper(eC.Event.ServiceName)+"_"+domain.AuthorFailed)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	eC.Transaction.SpanID = span.SpanContext().SpanID.String()
	eC.Transaction.TraceID = span.SpanContext().TraceID.String()

	event := eventbus.NewEvent(e.cfg.Service, eC.Event.EventType, eC.Event.Priority, eventbus.ProviderKafka, []byte(msg))
	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id":  eC.Transaction.ID,
			"root_id":         eC.Transaction.RootID,
			"span_id":         eC.Transaction.SpanID,
			"trace_id":        eC.Transaction.TraceID,
			"operation":       eC.Transaction.Operation,
			"snapshot":        eC.Transaction.Snapshot,
			"tracing_context": string(spanJSON),
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
	_, err = e.defaultCircuitBreaker("owner_failed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *AuthorSAGAKafkaEventBus) Created(ctx context.Context, author domain.Author) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Do any local low-volatile operation before any TCP/UDP connection
	authorJSON, err := json.Marshal(author)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"author", "author entity"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "author: created")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.AuthorCreated))

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	// Send domain event, spread aggregation side-effects to all required services
	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, authorJSON)
	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"tracing_context": string(spanJSON),
			"service":         event.ServiceName,
			"event_id":        event.ID,
			"event_type":      event.EventType,
			"priority":        event.Priority,
			"provider":        event.Provider,
			"dispatch_time":   event.DispatchTime,
		},
		BeforeSend: nil,
	}

	topic, err := eventbus.NewKafkaProducer(ctx, domain.AuthorCreated)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("created").Execute(func() (interface{}, error) {
		return nil, topic.Send(ctx, m)
	})

	return err
}

func (e *AuthorSAGAKafkaEventBus) BlobFailed(ctx context.Context, msg string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Owner/User verified, publish SERVICE_OWNER_VERIFIED
	ec, err := eventbus.ExtractContext(ctx)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"event", "event context"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "author: blob_failed")
	defer span.End()
	ctx = ctxT

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.BlobFailed))

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
	m := &pubsub.Message{
		Body: ev.Content,
		Metadata: map[string]string{
			"transaction_id":  ec.Transaction.ID,
			"root_id":         ec.Transaction.RootID,
			"span_id":         ec.Transaction.SpanID,
			"trace_id":        ec.Transaction.TraceID,
			"operation":       ec.Transaction.Operation,
			"snapshot":        ec.Transaction.Snapshot,
			"tracing_context": string(spanJSON),
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

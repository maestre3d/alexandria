package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"strings"
	"sync"
	"time"
)

type UserSAGAKafkaEvent struct {
	cfg *config.Kernel
	mu  *sync.Mutex
}

func NewUserSAGAKafkaEvent(cfg *config.Kernel) *UserSAGAKafkaEvent {
	return &UserSAGAKafkaEvent{
		cfg: cfg,
		mu:  new(sync.Mutex),
	}
}

func (e UserSAGAKafkaEvent) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "user_saga_kafka_" + action,
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

func (e *UserSAGAKafkaEvent) Verified(ctx context.Context) error {
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
	ctxT, span := trace.StartSpan(ctx, "identity: verified")
	defer span.End()

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", strings.ToUpper(eC.Event.ServiceName)+"_"+domain.OwnerVerified))

	// Prepare our span context to message metadata
	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	p, err := eventbus.NewKafkaProducer(ctxT, strings.ToUpper(eC.Event.ServiceName)+"_"+domain.OwnerVerified)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctxT)

	eC.Transaction.SpanID = span.SpanContext().SpanID.String()
	eC.Transaction.TraceID = span.SpanContext().TraceID.String()

	event := eventbus.NewEvent(e.cfg.Service, eC.Event.EventType, eC.Event.Priority, eventbus.ProviderKafka, []byte("user verified"))
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
	_, err = e.defaultCircuitBreaker("owner_verified").Execute(func() (interface{}, error) {
		return nil, p.Send(ctxT, m)
	})

	return err
}

func (e *UserSAGAKafkaEvent) Failed(ctx context.Context, msg string) error {
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
	ctxT, span := trace.StartSpan(ctx, "identity: failed")
	defer span.End()

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "send event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", strings.ToUpper(eC.Event.ServiceName)+"_"+domain.OwnerFailed))

	// Prepare our span context to message metadata
	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context"))
	}

	p, err := eventbus.NewKafkaProducer(ctxT, strings.ToUpper(eC.Event.ServiceName)+"_"+domain.OwnerFailed)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctxT)

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
		return nil, p.Send(ctxT, m)
	})

	return err
}

func (e *UserSAGAKafkaEvent) BlobFailed(ctx context.Context, msg string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// User failed to update, publish BLOB_FAILED
	eC, err := eventbus.ExtractContext(ctx)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"event", "event context"))
	}

	// Add tracing
	ctxT, span := trace.StartSpan(ctx, "identity: blob_failed")
	defer span.End()

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

	p, err := eventbus.NewKafkaProducer(ctxT, domain.BlobFailed)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctxT)

	eC.Transaction.SpanID = span.SpanContext().SpanID.String()
	eC.Transaction.TraceID = span.SpanContext().TraceID.String()

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderKafka, []byte(msg))
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

	_, err = e.defaultCircuitBreaker("blob_failed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctxT, m)
	})

	return err
}

package infrastructure

import (
	"context"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/sony/gobreaker"
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

	p, err := eventbus.NewKafkaProducer(ctx, strings.ToUpper(eC.Event.ServiceName)+"_"+domain.OwnerVerified)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	event := eventbus.NewEvent(e.cfg.Service, eC.Event.EventType, eC.Event.Priority, eventbus.ProviderKafka, []byte(""))
	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id": eC.Transaction.ID,
			"root_id":        eC.Transaction.RootID,
			"span_id":        eC.Transaction.SpanID,
			"trace_id":       eC.Transaction.TraceID,
			"operation":      eC.Transaction.Operation,
			"backup":         eC.Transaction.Backup,
			"service":        event.ServiceName,
			"event_id":       event.ID,
			"event_type":     event.EventType,
			"priority":       event.Priority,
			"provider":       event.Provider,
			"dispatch_time":  event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("owner_verified").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *UserSAGAKafkaEvent) Failed(ctx context.Context) error {
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

	p, err := eventbus.NewKafkaProducer(ctx, strings.ToUpper(eC.Event.ServiceName)+"_"+domain.OwnerFailed)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	event := eventbus.NewEvent(e.cfg.Service, eC.Event.EventType, eC.Event.Priority, eventbus.ProviderKafka, []byte(""))
	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"transaction_id": eC.Transaction.ID,
			"root_id":        eC.Transaction.RootID,
			"span_id":        eC.Transaction.SpanID,
			"trace_id":       eC.Transaction.TraceID,
			"operation":      eC.Transaction.Operation,
			"backup":         eC.Transaction.Backup,
			"service":        event.ServiceName,
			"event_id":       event.ID,
			"event_type":     event.EventType,
			"priority":       event.Priority,
			"provider":       event.Provider,
			"dispatch_time":  event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("owner_failed").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

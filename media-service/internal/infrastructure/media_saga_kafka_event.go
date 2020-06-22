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

	// Verify Author ID, publish AUTHOR_VERIFY
	eC, err := eventbus.ExtractContext(ctx)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"event", "event context"))
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

	event := eventbus.NewEvent(e.cfg.Service, eC.Event.EventType, eC.Event.Priority, eventbus.ProviderKafka, authorJSON)
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
	_, err = e.defaultCircuitBreaker("start_create").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

func (e *MediaSAGAKafkaEvent) Created(ctx context.Context, media domain.Media) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	mediaJSON, err := json.Marshal(media)
	if err != nil {
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"media", "media entity"))
	}

	event := eventbus.NewEvent(e.cfg.Service, eventbus.EventDomain, eventbus.PriorityLow, eventbus.ProviderKafka, mediaJSON)

	p, err := eventbus.NewKafkaProducer(ctx, domain.MediaCreated)
	if err != nil {
		return err
	}
	defer p.Shutdown(ctx)

	m := &pubsub.Message{
		Body: event.Content,
		Metadata: map[string]string{
			"service":       event.ServiceName,
			"event_id":      event.ID,
			"event_type":    event.EventType,
			"priority":      event.Priority,
			"provider":      event.Provider,
			"dispatch_time": event.DispatchTime,
		},
		BeforeSend: nil,
	}

	// Safe-call with circuit breaker pattern
	_, err = e.defaultCircuitBreaker("created").Execute(func() (interface{}, error) {
		return nil, p.Send(ctx, m)
	})

	return err
}

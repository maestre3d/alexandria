package bind

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"time"
)

type BlobEventConsumer struct {
	svc    usecase.BlobSagaInteractor
	logger log.Logger
	cfg    *config.Kernel
}

func NewBlobEventConsumer(svc usecase.BlobSagaInteractor, logger log.Logger, cfg *config.Kernel) *BlobEventConsumer {
	return &BlobEventConsumer{
		svc:    svc,
		logger: logger,
		cfg:    cfg,
	}
}

func (c BlobEventConsumer) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "blob_consumer_kafka_" + action,
		MaxRequests: 10,
		Interval:    10 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: nil,
	}

	return gobreaker.NewCircuitBreaker(st)
}

func extractContext(r *eventbus.Request) *eventbus.EventContext {
	return &eventbus.EventContext{
		Transaction: &eventbus.Transaction{
			ID:        r.Message.Metadata["transaction_id"],
			RootID:    r.Message.Metadata["root_id"],
			SpanID:    r.Message.Metadata["span_id"],
			TraceID:   r.Message.Metadata["trace_id"],
			Operation: r.Message.Metadata["operation"],
			Snapshot:  r.Message.Metadata["snapshot"],
		},
		Event: &eventbus.Event{
			Content:      r.Message.Body,
			ID:           r.Message.Metadata["event_id"],
			ServiceName:  r.Message.Metadata["service"],
			EventType:    r.Message.Metadata["event_type"],
			Priority:     r.Message.Metadata["priority"],
			Provider:     r.Message.Metadata["provider"],
			DispatchTime: r.Message.Metadata["dispatch_time"],
		},
	}
}

func (c *BlobEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	ctxC, _ := context.WithCancel(ctx)
	failedC, err := c.bindBlobFailed(ctxC, service)
	if err != nil {
		return err
	}

	s.AddConsumer(failedC)
	return nil
}

func (c *BlobEventConsumer) bindBlobFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	consumer, err := c.defaultCircuitBreaker("blob_failed").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.BlobFailed)
		if err != nil {
			return nil, err
		}

		return &eventbus.Consumer{
			MaxHandler: 10,
			Consumer:   sub,
			Handler:    c.onBlobFailed,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return consumer.(*eventbus.Consumer), nil
}

func (c *BlobEventConsumer) onBlobFailed(r *eventbus.Request) {
	eC := extractContext(r)

	// Get span context from message
	var traceCtx trace.SpanContext
	err := json.Unmarshal([]byte(r.Message.Metadata["tracing_context"]), &traceCtx)
	if err != nil {
		// If span is not valid, then create one from our current context
		rootSpan := trace.FromContext(r.Context)
		defer rootSpan.End()
		traceCtx = rootSpan.SpanContext()
	}

	// Start a new span with the parent span
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "blob: failed", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeInvalidArgument,
		Message: string(eC.Event.Content),
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.BlobFailed))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), eC)
	err = c.svc.Failed(ctxU, eC.Transaction.RootID, eC.Event.ServiceName, []byte(eC.Transaction.Snapshot))
	if err != nil {
		_ = level.Error(c.logger).Log("err", err)
		// If internal error, do nack
		if code := httputil.ErrorToCode(err); code == 500 {
			if r.Message.Nackable() {
				r.Message.Nack()
				return
			}
		}
	}

	r.Message.Ack()
}

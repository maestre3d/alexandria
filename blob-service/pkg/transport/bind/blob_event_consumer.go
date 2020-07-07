package bind

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/blob-service/internal/domain"
	"github.com/maestre3d/alexandria/blob-service/pkg/blob/usecase"
	"github.com/sony/gobreaker"
	"gocloud.dev/pubsub"
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

func injectEventContext(r *eventbus.Request) *eventbus.EventContext {
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
	sub, err := c.defaultCircuitBreaker("blob_failed").Execute(func() (interface{}, error) {
		return eventbus.NewKafkaConsumer(ctx, service, domain.BlobFailed)
	})
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub.(*pubsub.Subscription),
		Handler:    nil,
	}, nil
}

func (c *BlobEventConsumer) onBlobFailed(r *eventbus.Request) {
	eC := injectEventContext(r)

	ctxU := context.WithValue(r.Context, eventbus.EventContextKey("event"), eC)
	err := c.svc.Failed(ctxU, eC.Transaction.RootID, eC.Event.ServiceName, eC.Event.Content)
	if err != nil {
		// If internal error, do nack
		code := httputil.ErrorToCode(err)
		if code == 500 {
			if r.Message.Nackable() {
				r.Message.Nack()
			}

			_ = level.Error(c.logger).Log("err", err)
			return
		}
	}

	r.Message.Ack()
}

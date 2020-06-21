package bind

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
)

type UserEventConsumer struct {
	svc    usecase.UserSAGAInteractor
	logger log.Logger
	cfg    *config.Kernel
}

func NewUserEventConsumer(svc usecase.UserSAGAInteractor, logger log.Logger, cfg *config.Kernel) *UserEventConsumer {
	// TODO: Add instrumentation, Dist tracing with OpenTracing and Zipkin/Jaeger and Metrics with Prometheus w/ Grafana
	return &UserEventConsumer{
		svc:    svc,
		logger: logger,
		cfg:    cfg,
	}
}

func (c *UserEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	verifyBind, err := c.bindOwnerVerify(ctx, service)
	if err != nil {
		return err
	}

	s.AddConsumer(verifyBind)

	return nil
}

// Consumers / Binders
func (c *UserEventConsumer) bindOwnerVerify(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerVerify)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onOwnerVerify,
	}, nil
}

// Hooks / Handlers
func (c *UserEventConsumer) onOwnerVerify(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
	eC := &eventbus.EventContext{
		Transaction: &eventbus.Transaction{
			ID:        r.Message.Metadata["transaction_id"],
			RootID:    r.Message.Metadata["root_id"],
			SpanID:    r.Message.Metadata["span_id"],
			TraceID:   r.Message.Metadata["trace_id"],
			Operation: r.Message.Metadata["operation"],
			Backup:    r.Message.Metadata["backup"],
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

	ctxU := context.WithValue(r.Context, eventbus.EventContextKey("event"), eC)
	err := c.svc.Verify(ctxU, eC.Event.Content)
	if err != nil {
		// If internal error, do nack
		code := httputil.ErrorToCode(err)
		if code == 500 {
			if r.Message.Nackable() {
				r.Message.Nack()
			}
			_ = c.logger.Log("method", "identity.transport.event", "msg", "err", err.Error())
			return
		}
	}

	r.Message.Ack()
}

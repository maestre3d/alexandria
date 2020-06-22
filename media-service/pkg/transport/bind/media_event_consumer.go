package bind

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
)

type MediaEventConsumer struct {
	svc    usecase.MediaSAGAInteractor
	logger log.Logger
	cfg    *config.Kernel
}

func NewMediaEventConsumer(svc usecase.MediaSAGAInteractor, logger log.Logger, cfg *config.Kernel) *MediaEventConsumer {
	// TODO: Add instrumentation, Dist tracing with OpenTracing and Zipkin/Jaeger and Metrics with Prometheus w/ Grafana
	return &MediaEventConsumer{
		svc:    svc,
		logger: logger,
		cfg:    cfg,
	}
}

func (c *MediaEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	verifyBind, err := c.bindOwnerVerified(ctx, service)
	if err != nil {
		return err
	}

	failedBind, err := c.bindOwnerFailed(ctx, service)
	if err != nil {
		return err
	}

	aVerify, err := c.bindAuthorVerified(ctx, service)
	if err != nil {
		return err
	}

	aFailed, err := c.bindAuthorFailed(ctx, service)
	if err != nil {
		return err
	}

	s.AddConsumer(verifyBind)
	s.AddConsumer(failedBind)
	s.AddConsumer(aVerify)
	s.AddConsumer(aFailed)

	return nil
}

// Consumers / Binders
func (c *MediaEventConsumer) bindOwnerVerified(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerVerified)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onOwnerVerified,
	}, nil
}

func (c *MediaEventConsumer) bindOwnerFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerFailed)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onMediaFailed,
	}, nil
}

func (c *MediaEventConsumer) bindAuthorVerified(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.AuthorVerified)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onAuthorVerified,
	}, nil
}

func (c *MediaEventConsumer) bindAuthorFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.AuthorFailed)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onMediaFailed,
	}, nil
}

// Hooks / Handlers
func (c *MediaEventConsumer) onOwnerVerified(r *eventbus.Request) {
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

	// After owner validation, send AUTHOR_VERIFY event to validate authors now
	ctxU := context.WithValue(r.Context, eventbus.EventContextKey("event"), eC)
	err := c.svc.VerifyAuthor(ctxU, eC.Transaction.RootID)
	if err != nil {
		// If internal error, do nack
		code := httputil.ErrorToCode(err)
		if code == 500 {
			if r.Message.Nackable() {
				r.Message.Nack()
			}
			return
		}
	}

	// Always send acknowledge if operation succeed
	r.Message.Ack()
}

func (c *MediaEventConsumer) onAuthorVerified(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
	eC := &eventbus.EventContext{
		Transaction: &eventbus.Transaction{
			ID:        r.Message.Metadata["transaction_id"],
			RootID:    r.Message.Metadata["root_id"],
			SpanID:    r.Message.Metadata["span_id"],
			TraceID:   r.Message.Metadata["trace_id"],
			Operation: r.Message.Metadata["operation"],
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
	err := c.svc.Done(ctxU, eC.Transaction.RootID, eC.Transaction.Operation)
	if err != nil {
		// If internal error, do nack
		code := httputil.ErrorToCode(err)
		if code == 500 {
			if r.Message.Nackable() {
				r.Message.Nack()
			}
			return
		}
	}

	// Always send acknowledge if operation succeed
	r.Message.Ack()
}

func (c *MediaEventConsumer) onMediaFailed(r *eventbus.Request) {
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
	err := c.svc.Failed(ctxU, eC.Transaction.RootID, eC.Transaction.Operation, eC.Transaction.Backup)
	if err != nil {
		// If internal error, do nack
		code := httputil.ErrorToCode(err)
		if code == 500 {
			if r.Message.Nackable() {
				r.Message.Nack()
			}
			return
		}
	}

	r.Message.Ack()
}

package bind

import (
	"context"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
)

type AuthorEventConsumer struct {
	svc    usecase.AuthorSAGAInteractor
	logger log.Logger
}

func NewAuthorEventConsumer(svc usecase.AuthorSAGAInteractor, logger log.Logger) *AuthorEventConsumer {
	// TODO: Add instrumentation, Dist tracing with OpenTracing and Zipkin/Jaeger and Metrics with Prometheus w/ Grafana
	return &AuthorEventConsumer{
		svc:    svc,
		logger: logger,
	}
}

func (c *AuthorEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	aVerify, err := c.bindAuthorVerify(ctx, service)

	verifyBind, err := c.bindAuthorVerified(ctx, service)
	if err != nil {
		return err
	}

	failedBind, err := c.bindAuthorFailed(ctx, service)
	if err != nil {
		return err
	}

	s.AddConsumer(aVerify)
	s.AddConsumer(verifyBind)
	s.AddConsumer(failedBind)

	return nil
}

// Consumers / Binders

// Verifier listener
func (c *AuthorEventConsumer) bindAuthorVerify(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.AuthorVerify)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onAuthorVerify,
	}, nil
}

// Choreography-SAGA listeners / Foreign validations

func (c *AuthorEventConsumer) bindAuthorVerified(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerVerified)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onAuthorVerified,
	}, nil
}

func (c *AuthorEventConsumer) bindAuthorFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerFailed)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    c.onAuthorFailed,
	}, nil
}

// Hooks / Handlers

func (c *AuthorEventConsumer) onAuthorVerify(r *eventbus.Request) {
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
			_ = c.logger.Log("method", "author.transport.event", "msg", "err", err.Error())
			return
		}
	}

	r.Message.Ack()
}

func (c *AuthorEventConsumer) onAuthorVerified(r *eventbus.Request) {
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

func (c *AuthorEventConsumer) onAuthorFailed(r *eventbus.Request) {
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

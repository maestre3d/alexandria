package bind

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"time"
)

type AuthorEventConsumer struct {
	svc    usecase.AuthorSAGAInteractor
	logger log.Logger
}

func NewAuthorEventConsumer(svc usecase.AuthorSAGAInteractor, logger log.Logger) *AuthorEventConsumer {
	return &AuthorEventConsumer{
		svc:    svc,
		logger: logger,
	}
}

func (c AuthorEventConsumer) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "author_consumer_kafka_" + action,
		MaxRequests: 5,
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

func (c *AuthorEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	aVerify, err := c.bindAuthorVerify(ctx, service)
	if err != nil {
		return err
	}

	verifyBind, err := c.bindAuthorVerified(ctx, service)
	if err != nil {
		return err
	}

	failedBind, err := c.bindAuthorFailed(ctx, service)
	if err != nil {
		return err
	}

	blobU, err := c.bindBlobUploaded(ctx, service)
	if err != nil {
		return err
	}

	blobF, err := c.bindAuthorFailed(ctx, service)
	if err != nil {
		return err
	}

	s.AddConsumer(aVerify)
	s.AddConsumer(verifyBind)
	s.AddConsumer(failedBind)
	s.AddConsumer(blobU)
	s.AddConsumer(blobF)

	return nil
}

// Consumers / Binders

// Verifier listener
func (c *AuthorEventConsumer) bindAuthorVerify(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("author_verify").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.AuthorVerify)
		if err != nil {
			return nil, err
		}

		return sub, nil
	})
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub.(*pubsub.Subscription),
		Handler:    c.onAuthorVerify,
	}, nil
}

// Choreography-SAGA listeners / Foreign validations

func (c *AuthorEventConsumer) bindAuthorVerified(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("author_verified").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerVerified)
		if err != nil {
			return nil, err
		}

		return sub, nil
	})
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub.(*pubsub.Subscription),
		Handler:    c.onAuthorVerified,
	}, nil
}

func (c *AuthorEventConsumer) bindAuthorFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("author_failed").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerFailed)
		if err != nil {
			return nil, err
		}

		return sub, nil
	})
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub.(*pubsub.Subscription),
		Handler:    c.onAuthorFailed,
	}, nil
}

func (c *AuthorEventConsumer) bindBlobUploaded(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("blob_uploaded").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.BlobUploaded)
		if err != nil {
			return nil, err
		}

		return sub, nil
	})
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub.(*pubsub.Subscription),
		Handler:    c.onBlobUploaded,
	}, nil
}

func (c *AuthorEventConsumer) bindBlobRemoved(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("blob_removed").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.BlobRemoved)
		if err != nil {
			return nil, err
		}

		return sub, nil
	})
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub.(*pubsub.Subscription),
		Handler:    c.onBlobRemoved,
	}, nil
}

// Hooks / Handlers

func (c *AuthorEventConsumer) onAuthorVerify(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "author: verify", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeInvalidArgument,
		Message: string(eC.Event.Content),
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.AuthorVerify))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), eC)
	err = c.svc.Verify(ctxU, eC.Event.ServiceName, eC.Event.Content)
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

func (c *AuthorEventConsumer) onAuthorVerified(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "author: verified", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "event received",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.OwnerVerified))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), eC)
	err = c.svc.Done(ctxU, eC.Transaction.RootID, eC.Transaction.Operation)
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

	// Always send acknowledge if operation succeed
	r.Message.Ack()
}

func (c *AuthorEventConsumer) onAuthorFailed(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "author: failed", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeInvalidArgument,
		Message: string(eC.Event.Content),
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.OwnerFailed))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), eC)
	err = c.svc.Failed(ctxU, eC.Transaction.RootID, eC.Transaction.Operation, eC.Transaction.Snapshot)
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

func (c *AuthorEventConsumer) onBlobUploaded(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "author: blob_uploaded", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "event received",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.BlobUploaded))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), eC)
	err = c.svc.UpdatePicture(ctxU, eC.Transaction.RootID, eC.Event.Content)
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

func (c *AuthorEventConsumer) onBlobRemoved(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "author: blob_removed", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeInvalidArgument,
		Message: string(eC.Event.Content),
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.BlobUploaded))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), eC)
	err = c.svc.RemovePicture(ctxU, eC.Event.Content)
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

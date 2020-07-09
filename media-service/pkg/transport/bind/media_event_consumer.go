package bind

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"time"
)

type MediaEventConsumer struct {
	svc    usecase.MediaSAGAInteractor
	logger log.Logger
	cfg    *config.Kernel
}

func NewMediaEventConsumer(svc usecase.MediaSAGAInteractor, logger log.Logger, cfg *config.Kernel) *MediaEventConsumer {
	return &MediaEventConsumer{
		svc:    svc,
		logger: logger,
		cfg:    cfg,
	}
}

func (c MediaEventConsumer) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "media_consumer_kafka_" + action,
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
			Content:        r.Message.Body,
			TracingContext: r.Message.Metadata["tracing_context"],
			ID:             r.Message.Metadata["event_id"],
			ServiceName:    r.Message.Metadata["service"],
			EventType:      r.Message.Metadata["event_type"],
			Priority:       r.Message.Metadata["priority"],
			Provider:       r.Message.Metadata["provider"],
			DispatchTime:   r.Message.Metadata["dispatch_time"],
		},
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

	blobUp, err := c.bindBlobUploaded(ctx, service)
	if err != nil {
		return err
	}

	blobR, err := c.bindBlobRemoved(ctx, service)
	if err != nil {
		return err
	}

	s.AddConsumer(verifyBind)
	s.AddConsumer(failedBind)
	s.AddConsumer(aVerify)
	s.AddConsumer(aFailed)
	s.AddConsumer(blobUp)
	s.AddConsumer(blobR)

	return nil
}

// Consumers / Binders
func (c *MediaEventConsumer) bindOwnerVerified(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("owner_verified").Execute(func() (interface{}, error) {
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
		Handler:    c.onOwnerVerified,
	}, nil
}

func (c *MediaEventConsumer) bindOwnerFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("owner_verify").Execute(func() (interface{}, error) {
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
		Handler:    c.onMediaFailed,
	}, nil
}

func (c *MediaEventConsumer) bindAuthorVerified(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("author_verified").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.AuthorVerified)
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

func (c *MediaEventConsumer) bindAuthorFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := c.defaultCircuitBreaker("author_failed").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.AuthorFailed)
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
		Handler:    c.onMediaFailed,
	}, nil
}

func (c *MediaEventConsumer) bindBlobUploaded(ctx context.Context, service string) (*eventbus.Consumer, error) {
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

func (c *MediaEventConsumer) bindBlobRemoved(ctx context.Context, service string) (*eventbus.Consumer, error) {
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
func (c *MediaEventConsumer) onOwnerVerified(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
	ec := extractContext(r)

	// Get span context from message
	var traceCtx trace.SpanContext
	err := json.Unmarshal([]byte(ec.Event.TracingContext), &traceCtx)
	if err != nil {
		// If span is not valid, then create one from our current context
		rootSpan := trace.FromContext(r.Context)
		defer rootSpan.End()
		traceCtx = rootSpan.SpanContext()
	}

	// Start a new span with the parent span
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "media: user_verified", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "received event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.OwnerVerified))

	// After owner validation, send AUTHOR_VERIFY event to validate authors now
	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), ec)
	err = c.svc.VerifyAuthor(ctxU, ec.Transaction.RootID)
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

func (c *MediaEventConsumer) onAuthorVerified(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
	ec := extractContext(r)

	// Get span context from message
	var traceCtx trace.SpanContext
	err := json.Unmarshal([]byte(ec.Event.TracingContext), &traceCtx)
	if err != nil {
		// If span is not valid, then create one from our current context
		rootSpan := trace.FromContext(r.Context)
		defer rootSpan.End()
		traceCtx = rootSpan.SpanContext()
	}

	// Start a new span with the parent span
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "media: author_verified", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "received event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.AuthorVerified))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), ec)
	err = c.svc.Done(ctxU, ec.Transaction.RootID, ec.Transaction.Operation)
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

func (c *MediaEventConsumer) onMediaFailed(r *eventbus.Request) {
	// Wrap whole event for context propagation / OpenTracing-like
	ec := extractContext(r)

	// Get span context from message
	var traceCtx trace.SpanContext
	err := json.Unmarshal([]byte(ec.Event.TracingContext), &traceCtx)
	if err != nil {
		// If span is not valid, then create one from our current context
		rootSpan := trace.FromContext(r.Context)
		defer rootSpan.End()
		traceCtx = rootSpan.SpanContext()
	}

	// Start a new span with the parent span
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "media: failed", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeUnknown,
		Message: string(ec.Event.Content),
	})
	span.AddAttributes(trace.StringAttribute("event.name", ec.Transaction.Operation))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), ec)
	err = c.svc.Failed(ctxU, ec.Transaction.RootID, ec.Transaction.Operation, ec.Transaction.Snapshot)
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

func (c *MediaEventConsumer) onBlobUploaded(r *eventbus.Request) {
	ec := extractContext(r)

	var traceCtx trace.SpanContext
	err := json.Unmarshal([]byte(ec.Event.TracingContext), &traceCtx)
	if err != nil {
		rootSpan := trace.FromContext(r.Context)
		defer rootSpan.End()
		traceCtx = rootSpan.SpanContext()
	}

	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "media: blob_uploaded", traceCtx)
	defer span.End()

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "event received",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.BlobUploaded))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), ec)
	err = c.svc.UpdateStatic(ctxU, ec.Transaction.RootID, ec.Event.Content)
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

func (c *MediaEventConsumer) onBlobRemoved(r *eventbus.Request) {
	// Domain event (side-effects) does not use transactions
	ec := extractContext(r)

	var traceCtx trace.SpanContext
	err := json.Unmarshal([]byte(ec.Event.TracingContext), &traceCtx)
	if err != nil {
		rootSpan := trace.FromContext(r.Context)
		defer rootSpan.End()
		traceCtx = rootSpan.SpanContext()
	}

	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "media: blob_removed", traceCtx)
	defer span.End()

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "event received",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.BlobRemoved))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), ec)
	err = c.svc.RemoveStatic(ctxU, ec.Event.Content)
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

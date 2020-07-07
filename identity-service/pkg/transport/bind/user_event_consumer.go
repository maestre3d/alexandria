package bind

import (
	"context"
	"contrib.go.opencensus.io/exporter/zipkin"
	"encoding/json"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/sony/gobreaker"
	"go.opencensus.io/trace"
	"time"
)

type UserEventConsumer struct {
	svc    usecase.UserSAGAInteractor
	logger log.Logger
	cfg    *config.Kernel
}

func NewUserEventConsumer(svc usecase.UserSAGAInteractor, logger log.Logger, cfg *config.Kernel) *UserEventConsumer {
	// Set up trace exporters, this is meant to be done inside the transport injection, but currently
	// this service contains just event consumers, this is an exception and a custom implementation.
	// We avoid more injections inside this factory method because the trace exporter is used by just this transport
	// component (event consumer)

	// 1. Configure exporter to export traces to Zipkin.
	localEndpoint, err := openzipkin.NewEndpoint(cfg.Service, cfg.Tracing.ZipkinEndpoint)
	if err != nil {
		_ = level.Error(logger).Log("err", err.Error())
	}
	reporter := zipkinHTTP.NewReporter(cfg.Tracing.ZipkinHost)
	ze := zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(ze)

	// 2. Configure 100% sample rate, otherwise, few traces will be sampled.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return &UserEventConsumer{
		svc:    svc,
		logger: logger,
		cfg:    cfg,
	}
}

func (c UserEventConsumer) defaultCircuitBreaker(action string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        "user_consumer_kafka_" + action,
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
			ID:           r.Message.Metadata["event_id"],
			ServiceName:  r.Message.Metadata["service_name"],
			EventType:    r.Message.Metadata["event_type"],
			Content:      r.Message.Body,
			Priority:     r.Message.Metadata["priority"],
			Provider:     r.Message.Metadata["provider"],
			DispatchTime: r.Message.Metadata["dispatch_time"],
		},
	}
}

func (c *UserEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	ownerVerify, err := c.bindOwnerVerify(ctx, service)
	if err != nil {
		return err
	}
	s.AddConsumer(ownerVerify)

	blobUp, err := c.bindBlobUploaded(ctx, service)
	if err != nil {
		return err
	}
	s.AddConsumer(blobUp)

	blobRm, err := c.bindBlobRemoved(ctx, service)
	if err != nil {
		return err
	}
	s.AddConsumer(blobRm)

	return nil
}

// Consumers / Binders
func (c *UserEventConsumer) bindOwnerVerify(ctx context.Context, service string) (*eventbus.Consumer, error) {
	consumer, err := c.defaultCircuitBreaker("owner_verify").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerVerify)
		if err != nil {
			return nil, err
		}

		return &eventbus.Consumer{
			MaxHandler: 10,
			Consumer:   sub,
			Handler:    c.onOwnerVerify,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return consumer.(*eventbus.Consumer), nil
}

func (c *UserEventConsumer) bindBlobUploaded(ctx context.Context, service string) (*eventbus.Consumer, error) {
	consumer, err := c.defaultCircuitBreaker("blob_uploaded").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.BlobUploaded)
		if err != nil {
			return nil, err
		}

		return &eventbus.Consumer{
			MaxHandler: 10,
			Consumer:   sub,
			Handler:    c.onBlobUploaded,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return consumer.(*eventbus.Consumer), nil
}

func (c *UserEventConsumer) bindBlobRemoved(ctx context.Context, service string) (*eventbus.Consumer, error) {
	consumer, err := c.defaultCircuitBreaker("blob_removed").Execute(func() (interface{}, error) {
		sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.BlobRemoved)
		if err != nil {
			return nil, err
		}

		return &eventbus.Consumer{
			MaxHandler: 10,
			Consumer:   sub,
			Handler:    c.onBlobRemoved,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return consumer.(*eventbus.Consumer), nil
}

// Hooks / Handlers
func (c *UserEventConsumer) onOwnerVerify(r *eventbus.Request) {
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "identity: owner_verify", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "received event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.OwnerVerify))

	ctxU := context.WithValue(ctxT, eventbus.EventContextKey("event"), eC)
	err = c.svc.Verify(ctxU, eC.Event.Content)
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

func (c *UserEventConsumer) onBlobUploaded(r *eventbus.Request) {
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "identity: blob_uploaded", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "received event",
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

func (c *UserEventConsumer) onBlobRemoved(r *eventbus.Request) {
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
	ctxT, span := trace.StartSpanWithRemoteParent(r.Context, "identity: blob_removed", traceCtx)
	defer span.End()
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeOK,
		Message: "received event",
	})
	span.AddAttributes(trace.StringAttribute("event.name", domain.BlobRemoved))

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

package bind

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/httputil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
	"github.com/sony/gobreaker"
	"contrib.go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"time"
)

type UserEventConsumer struct {
	svc    usecase.UserSAGAInteractor
	logger log.Logger
	cfg    *config.Kernel
}

func NewUserEventConsumer(svc usecase.UserSAGAInteractor, logger log.Logger, cfg *config.Kernel) *UserEventConsumer {
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

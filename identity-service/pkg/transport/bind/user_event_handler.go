package bind

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
	"github.com/maestre3d/alexandria/identity-service/pkg/user/usecase"
	"gocloud.dev/pubsub"
	"strings"
)

type UserEventConsumer struct {
	svc    usecase.UserInteractor
	logger log.Logger
	cfg *config.Kernel
}

func NewUserEventConsumer(svc usecase.UserInteractor, logger log.Logger, cfg *config.Kernel) *UserEventConsumer {
	// TODO: Add instrumentation, Dist tracing with OpenTracing and Zipkin/Jaeger and Metrics with Prometheus w/ Grafana
	return &UserEventConsumer{
		svc:    svc,
		logger: logger,
		cfg: cfg,
	}
}

func (h *UserEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	verifyBind, err := h.bindOwnerVerify(ctx, service)
	if err != nil {
		return err
	}

	s.AddConsumer(verifyBind)

	return nil
}

// Consumers / Binders
func (h *UserEventConsumer) bindOwnerVerify(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.AuthorPending)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    h.onOwnerVerify,
	}, nil
}

// Hooks / Handlers
func (h *UserEventConsumer) onOwnerVerify(r *eventbus.Request) {
	e := &eventbus.Event{
		Content: 	  r.Message.Body,
		ServiceName:  r.Message.Metadata["service"],
		EventType:    r.Message.Metadata["event_type"],
		Priority:     r.Message.Metadata["priority"],
		Provider:     r.Message.Metadata["provider"],
	}

	t := &eventbus.Transaction{
		ID:        r.Message.Metadata["transaction_id"],
		RootID:    r.Message.Metadata["root_id"],
		SpanID:    "",
		TraceID:   "",
		Operation: r.Message.Metadata["operation"],
		Backup:    r.Message.Metadata["backup"],
		Code:      0,
		Message:   "",
	}

	// owners contains just an array with users id's string
	var owners []string
	err := json.Unmarshal(e.Content, &owners)
	if err != nil {
		// TODO: Send failed event
		t.Code = 400
		t.Message = exception.InvalidFieldFormat.Error()
		err = h.ownerFailed(r.Context, e, t)
		if err != nil {
			// Error while connecting to broker, do not ack
			if r.Message.Nackable() {
				r.Message.Nack()
			}
			return
		}
	}

	// Verify
	for _, ownerID := range owners {
		_, err := h.svc.Get(r.Context, ownerID)
		if err != nil {
			// TODO: Send failed event
			t.Code = 404
			t.Message = fmt.Sprintf("%s: identity %s not found", exception.EntityNotFound.Error(), ownerID)
			err = h.ownerFailed(r.Context, e, t)
			if err != nil {
				// Error while connecting to broker, do not ack
				if r.Message.Nackable() {
					r.Message.Nack()
				}
			}
			return
		}
	}

	// TODO: Send verified event
	err = h.ownerVerified(r.Context, e, t)
	if err != nil {
		// Error while connecting to broker, do not ack
		if r.Message.Nackable() {
			r.Message.Nack()
		}
		return
	}

	r.Message.Ack()
}

func (h *UserEventConsumer) ownerFailed(ctx context.Context, e *eventbus.Event, t *eventbus.Transaction) error {
	// Identity not found, publish AUTHOR_OWNER_FAILED
	// This is supposed to be inside identity's use case event bus implementation
	topic, err := eventbus.NewKafkaProducer(ctx, fmt.Sprintf("%s_%s", strings.ToUpper(e.ServiceName), domain.OwnerFailed))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	h.logger.Log("msg", fmt.Sprintf("%+v", t))

	// Error event struct
	mJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}

	m := &pubsub.Message{
		Body: mJSON,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"operation":      t.Operation,
			"backup":		  t.Backup,
			"service":        strings.ToLower(h.cfg.Service),
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       eventbus.ProviderKafka,
		},
		BeforeSend: nil,
	}

	_ = h.logger.Log("method", "user.transport.owner_failed", "msg", fmt.Sprintf("%s_%s event published", strings.ToUpper(e.ServiceName),
		domain.OwnerFailed))

	return topic.Send(ctx, m)
}

func (h *UserEventConsumer) ownerVerified(ctx context.Context, e *eventbus.Event, t *eventbus.Transaction) error {
	// Identity not found, publish AUTHOR_OWNER_VERIFIED
	// This is supposed to be inside identity's use case event bus implementation
	topic, err := eventbus.NewKafkaProducer(ctx, fmt.Sprintf("%s_%s", strings.ToUpper(e.ServiceName), domain.OwnerVerified))
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	h.logger.Log("msg", fmt.Sprintf("%+v", t))

	// Error event struct
	mJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}

	m := &pubsub.Message{
		Body: mJSON,
		Metadata: map[string]string{
			"transaction_id": t.ID,
			"operation":      t.Operation,
			"service":        strings.ToLower(h.cfg.Service),
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       eventbus.ProviderKafka,
		},
		BeforeSend: nil,
	}

	_ = h.logger.Log("method", "user.transport.owner_verified", "msg", fmt.Sprintf("%s_%s event published", strings.ToUpper(e.ServiceName),
		domain.OwnerVerified))

	return topic.Send(ctx, m)
}
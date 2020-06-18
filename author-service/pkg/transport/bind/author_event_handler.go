package bind

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
)

type AuthorEventConsumer struct {
	svc    usecase.AuthorInteractor
	logger log.Logger
}

func NewAuthorEventConsumer(svc usecase.AuthorInteractor, logger log.Logger) *AuthorEventConsumer {
	// TODO: Add instrumentation, Dist tracing with OpenTracing and Zipkin/Jaeger and Metrics with Prometheus w/ Grafana
	return &AuthorEventConsumer{
		svc:    svc,
		logger: logger,
	}
}

func (h *AuthorEventConsumer) SetBinders(s *eventbus.Server, ctx context.Context, service string) error {
	verifyBind, err := h.bindAuthorVerified(ctx, service)
	if err != nil {
		return err
	}

	failedBind, err := h.bindAuthorFailed(ctx, service)
	if err != nil {
		return err
	}

	s.AddConsumer(verifyBind)
	s.AddConsumer(failedBind)

	return nil
}

// Consumers / Binders
func (h *AuthorEventConsumer) bindAuthorVerified(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerVerified)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    h.onAuthorVerified,
	}, nil
}

func (h *AuthorEventConsumer) bindAuthorFailed(ctx context.Context, service string) (*eventbus.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerFailed)
	if err != nil {
		return nil, err
	}

	return &eventbus.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    h.onAuthorFailed,
	}, nil
}

// Hooks / Handlers
func (h *AuthorEventConsumer) onAuthorVerified(r *eventbus.Request) {
	t := &eventbus.Transaction{}
	err := json.Unmarshal(r.Message.Body, t)
	if err != nil {
		// TODO: Rollback if malformation happens, ack message
		if r.Message.Nackable() {
			r.Message.Nack()
		}
		return
	}

	err = h.svc.Done(r.Context, t.RootID, t.Operation)
	if err != nil {
		// If not found, then acknowledge message
		if !errors.Is(err, exception.EntityNotFound) {
			if r.Message.Nackable() {
				r.Message.Nack()
			}
			return
		}
	}

	// Always send acknowledge if operation succeed
	r.Message.Ack()
}

func (h *AuthorEventConsumer) onAuthorFailed(r *eventbus.Request) {
	t := &eventbus.Transaction{}
	err := json.Unmarshal(r.Message.Body, t)
	if err != nil {
		// TODO: Rollback if malformation happens, ack message
		if r.Message.Nackable() {
			r.Message.Nack()
		}
		return
	}

	err = h.svc.Failed(r.Context, t.RootID, t.Operation, t.Backup)
	if err != nil {
		// If not found, then acknowledge message
		if !errors.Is(err, exception.EntityNotFound) {
			if r.Message.Nackable() {
				r.Message.Nack()
			}
			return
		}

	}

	r.Message.Ack()
}

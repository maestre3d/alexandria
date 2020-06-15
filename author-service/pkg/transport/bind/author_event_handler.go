package bind

import (
	"context"
	"encoding/json"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/maestre3d/alexandria/author-service/pkg/author/usecase"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/event"
)

type AuthorEventHandler struct {
	svc    usecase.AuthorInteractor
	logger log.Logger
}

func NewAuthorEventHandler(svc usecase.AuthorInteractor, logger log.Logger) *AuthorEventHandler {
	// TODO: Add instrumentation, Dist tracing with OpenTracing and Zipkin/Jaeger and Metrics with Prometheus w/ Grafana
	return &AuthorEventHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *AuthorEventHandler) SetBinders(s *event.Server, ctx context.Context, service string) error {
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
func (h *AuthorEventHandler) bindAuthorVerified(ctx context.Context, service string) (*event.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerVerified)
	if err != nil {
		return nil, err
	}

	return &event.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    h.onAuthorVerified,
	}, nil
}

func (h *AuthorEventHandler) bindAuthorFailed(ctx context.Context, service string) (*event.Consumer, error) {
	sub, err := eventbus.NewKafkaConsumer(ctx, service, domain.OwnerFailed)
	if err != nil {
		return nil, err
	}

	return &event.Consumer{
		MaxHandler: 10,
		Consumer:   sub,
		Handler:    h.onAuthorFailed,
	}, nil
}

// Hooks / Handlers
func (h *AuthorEventHandler) onAuthorVerified(r *event.Request) error {
	tr := &event.Transaction{}
	err := json.Unmarshal(r.Message.Body, tr)
	if err != nil {
		return err
	}

	return h.svc.Done(r.Context, tr.EntityID, r.Message.Metadata["operation"])
}

func (h *AuthorEventHandler) onAuthorFailed(r *event.Request) error {
	tr := &event.Transaction{}
	err := json.Unmarshal(r.Message.Body, tr)
	if err != nil {
		return err
	}

	return h.svc.Failed(r.Context, tr.EntityID, r.Message.Metadata["operation"], "")
}

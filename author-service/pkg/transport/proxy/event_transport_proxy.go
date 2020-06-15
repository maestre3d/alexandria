package proxy

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/maestre3d/alexandria/author-service/pkg/transport/event"
)

type Consumer interface {
	SetBinders(s *event.Server, ctx context.Context, service string) error
}

type Event struct {
	Server    *event.Server
	ctx       context.Context
	cfg       *config.Kernel
	consumers []Consumer
}

func NewEvent(ctx context.Context, cfg *config.Kernel, consumers ...Consumer) (*Event, func(), error) {
	ctxS, cancel := context.WithCancel(ctx)

	srv := event.NewServer(ctxS, cancel)
	clean := func() {
		srv.Close()
	}

	proxy := &Event{srv, ctx, cfg, consumers}

	err := proxy.mapRoutes()
	if err != nil {
		return nil, nil, err
	}

	return proxy, clean, nil
}

func (e *Event) mapRoutes() error {
	for _, consumer := range e.consumers {
		err := consumer.SetBinders(e.Server, e.ctx, e.cfg.Service)
		if err != nil {
			return err
		}
	}

	return nil
}

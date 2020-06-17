package proxy

import (
	"context"
	"github.com/alexandria-oss/core/config"
	"github.com/alexandria-oss/core/eventbus"
)

type Consumer interface {
	SetBinders(s *eventbus.Server, ctx context.Context, service string) error
}

type Event struct {
	Server    *eventbus.Server
	ctx       context.Context
	cfg       *config.Kernel
	consumers []Consumer
}

func NewEvent(ctx context.Context, cfg *config.Kernel, consumers ...Consumer) (*Event, func(), error) {
	ctxS, _ := context.WithCancel(ctx)

	srv := eventbus.NewServer(ctxS)
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

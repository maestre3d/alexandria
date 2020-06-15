package event

import (
	"context"
	"gocloud.dev/pubsub"
	"sync"
)

type Transaction struct {
	EntityID string `json:"entity_id"`
	Code     int    `json:"code"`
	Message  string `json:"message"`
}

type Request struct {
	Context context.Context
	Message *pubsub.Message
}

type HandlerFunc func(*Request) error

type Consumer struct {
	MaxHandler int
	Consumer   *pubsub.Subscription
	Handler    HandlerFunc
	cancelCtx  context.CancelFunc
}

func (s *Consumer) serve(ctx context.Context) {
	defer func() {
		_ = s.Consumer.Shutdown(ctx)
	}()

	// Loop on received messages. We can use a channel as a semaphore to limit how
	// many goroutines we have active at a time as well as wait on the goroutines
	// to finish before exiting.
	sem := make(chan struct{}, s.MaxHandler)
recvLoop:
	for {
		msg, err := s.Consumer.Receive(ctx)
		if err != nil {
			// Errors from Receive indicate that Receive will no longer succeed.
			s.cancelCtx()
			break
		}

		// Wait if there are too many active handle goroutines and acquire the
		// semaphore. If the context is canceled, stop waiting and start shutting
		// down.
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break recvLoop
		}

		// Handle the message in a new goroutine.
		go func() {
			defer func() { <-sem }() // Release the semaphore.

			// Do work based on the message
			ctxHdl, _ := context.WithCancel(ctx)
			err = s.Handler(&Request{
				Context: ctxHdl,
				Message: msg,
			})
			// err = validateOwners(ctx, author, msg.Metadata["transaction_id"])
			if err != nil {
				return
			}

			// Just send acknowledgement to broker if handler throws nil
			msg.Ack() // Messages must always be acknowledged with Ack.
		}()
	}

	// We're no longer receiving messages. Wait to finish handling any
	// unacknowledged messages by totally acquiring the semaphore.
	for n := 0; n < s.MaxHandler; n++ {
		sem <- struct{}{}
	}
}

type Server struct {
	Consumers   []*Consumer
	rootContext context.Context
	mtx         *sync.Mutex
	cancelCtx   context.CancelFunc
}

func NewServer(ctx context.Context, cancel context.CancelFunc, cs ...*Consumer) *Server {
	return &Server{
		Consumers:   cs,
		rootContext: ctx,
		mtx:         new(sync.Mutex),
		cancelCtx:   cancel,
	}
}

func (s *Server) AddConsumer(c *Consumer) {
	s.Consumers = append(s.Consumers, c)
}

func (s *Server) Serve() error {
	for _, c := range s.Consumers {
		ctxSub, cancel := context.WithCancel(s.rootContext)
		c.cancelCtx = cancel
		go c.serve(ctxSub)
	}

	select {
	case <-s.rootContext.Done():
		return nil
	}
}

func (s *Server) Close() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.rootContext.Done()
}

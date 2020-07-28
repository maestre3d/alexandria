package eventutil

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/eapache/go-resiliency/retrier"
	"go.opencensus.io/trace"
	"gocloud.dev/pubsub"
	"strings"
	"time"
)

type EventAggregate struct {
	Name    string
	Prefix  string
	Topic   *pubsub.Topic
	Message *pubsub.Message
}

// PublishResilientEvent make a call to an message broker safely
func PublishResilientEvent(ctx context.Context, ag EventAggregate) error {
	// Add prefix if not empty
	commandName := ""
	if ag.Prefix != "" {
		commandName = ag.Prefix + "_"
	}
	commandName += ag.Name

	// Circuit Breaker wrapping
	output := make(chan bool, 1)
	errors := hystrix.Go(strings.ToLower(commandName), func() error {
		// Retry policy wrapping
		r := retrier.New(retrier.ConstantBackoff(3, 150*time.Millisecond), nil)
		err := r.Run(func() error {
			err := ag.Topic.Send(ctx, ag.Message)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		output <- true
		return nil
	}, func(err error) error {
		// Add logging if required
		return err
	})

	select {
	case <-output:
		return nil
	case err := <-errors:
		return err
	}
}

// Parse OpenCensus span context to JSON safely
func SpanCtxToJSON(ctx context.Context) ([]byte, error) {
	span := trace.FromContext(ctx)
	defer span.End()

	spanJSON, err := json.Marshal(span.SpanContext())
	if err != nil {
		return nil, exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"tracing_context", "span context object"))
	}

	return spanJSON, nil
}

// Generate required event metadata
func GenerateEventMetadata(event eventbus.Event) map[string]string {
	return map[string]string{
		"event_id":        event.ID,
		"tracing_context": event.TracingContext,
		"service":         event.ServiceName,
		"event_type":      event.EventType,
		"priority":        event.Priority,
		"provider":        event.Provider,
		"dispatch_time":   event.DispatchTime,
	}
}

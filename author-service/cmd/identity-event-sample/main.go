package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"github.com/oklog/run"
	"gocloud.dev/pubsub"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "gocloud.dev/pubsub/awssnssqs"
)

func main() {
	var g run.Group
	{
		// Set up signal bind
		var (
			cancelInterrupt = make(chan struct{})
			c               = make(chan os.Signal, 2)
			ctx             = context.Background()
		)
		defer func() {
			close(c)
			ctx.Done()
		}()

		g.Add(func() error {
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				err := listenSAGA(ctx)
				log.Print(err)
			}()

			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	_ = g.Run()
}

func listenSAGA(ctx context.Context) error {
	errChan := make(chan error)
	defer close(errChan)
	go func() {
		err := onAuthorCreated(ctx)
		if err != nil {
			log.Print(err)
		}

		errChan <- err
	}()
	select {
	case err := <-errChan:
		return err
	}
}

func onAuthorCreated(ctx context.Context) error {
	sub, err := pubsub.OpenSubscription(ctx, "awssqs://sqs.us-east-1.amazonaws.com/824699638576/alexandria_author_created-identity?region=us-east-1")
	if err != nil {
		return err
	}
	defer sub.Shutdown(ctx)

	mChan := make(chan *pubsub.Message)
	eChan := make(chan error)
	defer func() {
		close(mChan)
		close(eChan)
	}()
	go func() {
		for {
			log.Print("hook onAuthorCreated listening")
			m, err := sub.Receive(ctx)
			if err != nil {
				eChan <- err
				break
			}

			mChan <- m
		}
	}()

	for {
		select {
		case m := <-mChan:
			m.Ack()

			author := new(domain.Author)
			err := json.Unmarshal(m.Body, author)
			if err != nil {
				eChan <- err
			}

			log.Printf("%v", author)

			err = validateOwners(ctx, author, m.Metadata["transaction_id"])
			if err != nil {
				eChan <- err
			}

		case err = <-eChan:
			return err
		}
	}
}

/* Use case */

func validateOwners(ctx context.Context, author *domain.Author, transactionID string) error {
	for _, owner := range author.Owners {
		err := getIdentity(owner.ID)
		if err != nil {
			// Identity not found, publish ALEXANDRIA_AUTHOR_IDENTITY_NOT_FOUND
			// This is supposed to be inside identity's use case event bus implementation
			topic, err := pubsub.OpenTopic(ctx, "awssns:///arn:aws:sns:us-east-1:824699638576:ALEXANDRIA_AUTHOR_IDENTITY_NOT_FOUND?region=us-east-1")
			if err != nil {
				return err
			}
			defer topic.Shutdown(ctx)

			// Error event struct
			mJSON, err := json.Marshal(&struct {
				Message  string `json:"message"`
				Code     int32  `json:"code"`
				EntityID string `json:"entity_id"`
			}{
				fmt.Sprintf("%s: identity %s not found", exception.EntityNotFound.Error(), owner.ID),
				404,
				author.ExternalID,
			})

			e := eventbus.NewEvent("identity", eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, mJSON, true)
			m := &pubsub.Message{
				Body: e.Content,
				Metadata: map[string]string{
					"transaction_id": transactionID,
					"service":        e.ServiceName,
					"type":           e.EventType,
					"priority":       e.Priority,
				},
				BeforeSend: nil,
			}

			return topic.Send(ctx, m)
		}
	}

	// All identities found, Publish ALEXANDRIA_AUTHOR_IDENTITY_VERIFIED
	// This is supposed to be inside identity's use case event bus implementation
	topic, err := pubsub.OpenTopic(ctx, "awssns:///arn:aws:sns:us-east-1:824699638576:ALEXANDRIA_AUTHOR_IDENTITY_VERIFIED?region=us-east-1")
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Transaction event struct
	mJSON, err := json.Marshal(&struct {
		Message  string `json:"message"`
		Code     int32  `json:"code"`
		EntityID string `json:"entity_id"`
	}{
		"",
		200,
		author.ExternalID,
	})

	e := eventbus.NewEvent("identity", eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, mJSON, true)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": transactionID,
			"service":        e.ServiceName,
			"type":           e.EventType,
			"priority":       e.Priority,
		},
		BeforeSend: nil,
	}

	return topic.Send(ctx, m)
}

/* Helpers/Mock */

func getIdentity(id string) error {
	identityDB := make([]*domain.Owner, 0)
	identityDB = append(identityDB, &domain.Owner{
		ID: "a0838eef-42dd-40b2-87bd-9dde180a3cae",
	})

	found := false
	for _, identity := range identityDB {
		if id == identity.ID {
			found = true
		}
	}

	if !found {
		return errors.New("identity not found")
	}

	return nil
}

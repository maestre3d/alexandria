package usecasemock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandria-oss/core/eventbus"
	"github.com/alexandria-oss/core/exception"
	"github.com/maestre3d/alexandria/author-service/internal/domain"
	"gocloud.dev/pubsub"
	"log"
)

const (
	ServiceName = "identity"
)

/* Use case */

func ValidateOwners(ctx context.Context, author *domain.Author, transactionID, operation string) error {
	err := getIdentity(author.OwnerID)
	if err != nil {
		// Identity not found, publish AUTHOR_OWNER_FAILED
		// This is supposed to be inside identity's use case event bus implementation
		topic, err := eventbus.NewKafkaProducer(ctx, domain.OwnerFailed)
		if err != nil {
			return err
		}
		defer topic.Shutdown(ctx)

		// Error event struct
		mJSON, err := json.Marshal(&eventbus.Transaction{
			ID:        transactionID,
			RootID:    author.ExternalID,
			Operation: operation,
			Code:      404,
			Message:   fmt.Sprintf("%s: identity %s not found", exception.EntityNotFound.Error(), author.OwnerID),
		})

		e := eventbus.NewEvent(ServiceName, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, mJSON)
		m := &pubsub.Message{
			Body: e.Content,
			Metadata: map[string]string{
				"transaction_id": transactionID,
				"operation":      operation,
				"service":        e.ServiceName,
				"event_type":     e.EventType,
				"priority":       e.Priority,
				"provider":       e.Provider,
			},
			BeforeSend: nil,
		}

		log.Print("AUTHOR_OWNER_FAILED event dispatched")
		return topic.Send(ctx, m)
	}

	// All identities found, Publish AUTHOR_OWNER_VERIFIED
	// This is supposed to be inside identity's use case event bus implementation
	topic, err := eventbus.NewKafkaProducer(ctx, domain.OwnerVerified)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	// Transaction event struct
	mJSON, err := json.Marshal(&eventbus.Transaction{
		ID:        transactionID,
		RootID:    author.ExternalID,
		SpanID:    "",
		TraceID:   "",
		Operation: operation,
		Backup:    "",
		Code:      200,
		Message:   "author " + author.ExternalID + " verified",
	})

	e := eventbus.NewEvent(ServiceName, eventbus.EventIntegration, eventbus.PriorityHigh, eventbus.ProviderAWS, mJSON)
	m := &pubsub.Message{
		Body: e.Content,
		Metadata: map[string]string{
			"transaction_id": transactionID,
			"operation":      operation,
			"service":        e.ServiceName,
			"event_type":     e.EventType,
			"priority":       e.Priority,
			"provider":       e.Provider,
		},
		BeforeSend: nil,
	}

	log.Print("AUTHOR_OWNER_VERIFIED event dispatched")
	return topic.Send(ctx, m)
}

/* Helpers/Mock */

func getIdentity(id string) error {
	identityDB := make([]*domain.Owner, 0)
	identityDB = append(identityDB, &domain.Owner{
		ID: "a0838eef-42dd-40b2-87bd-9dde180a3cae",
	})

	for _, identity := range identityDB {
		if id == identity.ID {
			return nil
		}
	}

	return errors.New("identity not found")
}

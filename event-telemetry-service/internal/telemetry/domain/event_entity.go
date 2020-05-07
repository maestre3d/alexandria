package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// EventEntity Represents an event log record
/*
*	Service Name = Service who dispatched the event
*	Transaction ID = Distributed transaction ID *Only for SAGA pattern
*	Event Type = Type of the event dispatched (integration or domain)
*	Content = Message body, mostly bytes or JSON-into string
*	Importance = Event's importance
*	Provider = Message Broker Provider (Kafka, RabbitMQ)
*	Dispatch Time = Event's dispatching timestamp
 */
type EventEntity struct {
	ID            string  `json:"event_id" validate:"required" docstore:"event_id"`
	ServiceName   string  `json:"service_name" validate:"required" docstore:"service_name"`
	TransactionID *string `json:"transaction_id,omitempty" docstore:"transaction_id,omitempty"`
	EventType     string  `json:"event_type" validate:"required" docstore:"event_type"`
	Content       string  `json:"content" validate:"required" docstore:"content"`
	Importance    string  `json:"importance" validate:"required" docstore:"importance"`
	Provider      string  `json:"provider" validate:"required" docstore:"provider"`
	DispatchTime  int64   `json:"dispatch_time" validate:"required" docstore:"dispatch_time"`
}

// NewEventEntity Create a new telemetry entity using domain rules
func NewEventEntity(serviceName, transactionID, eventType, content, importance, provider string) *EventEntity {
	var transaction *string

	if transactionID != "" {
		transaction = &transactionID
	} else {
		transaction = nil
	}

	return &EventEntity{
		ID:            uuid.New().String(),
		ServiceName:   strings.ToUpper(serviceName),
		TransactionID: transaction,
		EventType:     eventType,
		Content:       content,
		Importance:    importance,
		Provider:      strings.ToUpper(provider),
		DispatchTime:  time.Now().UnixNano() / 1000000,
	}
}

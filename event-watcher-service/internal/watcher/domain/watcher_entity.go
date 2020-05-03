package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// WatcherEntity Represents an event log record
/*
*	Service Name = Service who dispatched the event
*	Transaction ID = Distributed transaction ID *Only for SAGA pattern
*	Event Type = Type of the event dispatched (integration or domain)
*	Content = Message body, mostly bytes or JSON-into string
*	Importance = Event's importance
*	Provider = Message Broker Provider (Kafka, RabbitMQ)
*	Dispatch Time = Event's dispatching timestamp
 */
type WatcherEntity struct {
	ID            string    `json:"id" validate:"required"`
	ServiceName   string    `json:"service_name" validate:"required"`
	TransactionID *string   `json:"transaction_id,omitempty"`
	EventType     string    `json:"event_type" validate:"required"`
	Content       string    `json:"content" validate:"required"`
	Importance    string    `json:"importance" validate:"required"`
	Provider      string    `json:"provider" validate:"required"`
	DispatchTime  time.Time `json:"dispatch_time" validate:"required"`
}

func NewWatcherEntity(serviceName, transactionID, eventType, content, importance, provider string) *WatcherEntity {
	var transaction *string

	if transactionID != "" {
		transaction = &transactionID
	} else {
		transaction = nil
	}

	return &WatcherEntity{
		ID:            uuid.New().String(),
		ServiceName:   strings.ToUpper(serviceName),
		TransactionID: transaction,
		EventType:     eventType,
		Content:       content,
		Importance:    importance,
		Provider:      strings.ToUpper(provider),
		DispatchTime:  time.Now(),
	}
}

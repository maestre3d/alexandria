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
	ID            string    `json:"watcher_id" validate:"required" docstore:"watcher_id"`
	ServiceName   string    `json:"service_name" validate:"required" docstore:"service_name"`
	TransactionID *string   `json:"transaction_id,omitempty" docstore:"transaction_id,omitempty"`
	EventType     string    `json:"event_type" validate:"required" docstore:"event_type"`
	Content       string    `json:"content" validate:"required" docstore:"content"`
	Importance    string    `json:"importance" validate:"required" docstore:"importance"`
	Provider      string    `json:"provider" validate:"required" docstore:"provider"`
	DispatchTime  time.Time `json:"dispatch_time" validate:"required" docstore:"dispatch_time"`
}

// NewWatcherEntity Create a new watcher entity using domain rules
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

package models

import "time"

const (
	PaymentCreatedEventTopic = "payments.created"
	WalletDebitEventTopic    = "wallet.debit.requested"
	PaymentsDLQTopic         = "payments.dlq"
)

type PaymentCreatedEvent struct {
	ID         string    `json:"id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	Method     string    `json:"method"`
	CustomerID string    `json:"customer_id"`
	TraceID    string    `json:"trace_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type WalletDebitRequestedEvent struct {
	PaymentID string  `json:"payment_id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
	TraceID   string  `json:"trace_id"`
}

type DLQMessage struct {
	OriginalTopic string    `json:"original_topic"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	Timestamp     time.Time `json:"timestamp"`
	Attempts      int       `json:"attempts"`
}

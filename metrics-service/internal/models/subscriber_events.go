package models

import "time"

const (
	TopicPaymentsCreated      = "payments.created"
	TopicPaymentsChecked      = "payments.checked"
	TopicWalletResponse       = "wallet.response"
	TopicWalletDebitRequested = "wallet.debit.requested"
)

type FraudCheckEvent struct {
	ID        string    `json:"id"`
	TraceID   string    `json:"trace_id"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

type WalletResponseEvent struct {
	PaymentID string  `json:"payment_id"`
	UserID    string  `json:"user_id"`
	Status    string  `json:"status"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}

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

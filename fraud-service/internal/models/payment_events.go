package models

import "time"

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

package models

import "time"

const (
	PaymentStatusApproved = "APPROVED"
	PaymentStatusDeclined = "DECLINED"

	TopicPaymentChecked = "payments.checked"
	PaymentsDLQTopic    = "payments.dlq"
)

type FraudCheckEvent struct {
	ID        string    `json:"id"`
	TraceID   string    `json:"trace_id"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

type DLQMessage struct {
	OriginalTopic string    `json:"original_topic"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	Timestamp     time.Time `json:"timestamp"`
	Attempts      int       `json:"attempts"`
}

package models

import "time"

const (
	FraudTopic2Subscribe  string = "payments.checked"
	WalletTopic2Subscribe string = "wallet.funds.verified"

	PaymentStatusApproved = "APPROVED"
	PaymentStatusDeclined = "DECLINED"
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

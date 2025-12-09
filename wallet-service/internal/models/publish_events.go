package models

import "time"

type WalletStatus string

const (
	WalletStatusApproved WalletStatus = "APPROVED"
	WalletStatusDeclined WalletStatus = "DECLINED"

	WalletResponseTopic = "wallet.response"
	WalletDLQTopic      = "wallet.dlq"
)

type WalletResponseEvent struct {
	PaymentID string       `json:"payment_id"`
	UserID    string       `json:"user_id"`
	Status    WalletStatus `json:"status"`
	Amount    float64      `json:"amount"`
	Reason    string       `json:"reason"`
}

type DLQMessage struct {
	OriginalTopic string    `json:"original_topic"`
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	Timestamp     time.Time `json:"timestamp"`
	Attempts      int       `json:"attempts"`
}

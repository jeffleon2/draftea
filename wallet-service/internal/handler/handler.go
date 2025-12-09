package handler

import (
	"context"
	"encoding/json"

	"github.com/jeffleon2/draftea-wallet-service/internal/models"
	"github.com/sirupsen/logrus"
)

// WalletServiceIn defines the interface for wallet business logic operations.
type WalletServiceIn interface {
	ValidateFunds(ctx context.Context, event models.PaymentCreatedEvent) error
	DebitBalance(ctx context.Context, userID string, amount float64) error
}

// WalletHandler processes Kafka events for wallet operations.
// It handles funds verification and balance debit requests.
type WalletHandler struct {
	WalletService WalletServiceIn
}

// Wallet creates a new WalletHandler with the provided wallet service.
func Wallet(s WalletServiceIn) *WalletHandler {
	return &WalletHandler{
		WalletService: s,
	}
}

// Handler processes Kafka events for wallet operations based on the topic.
// It handles two types of events:
//   - wallet.debit.requested: Executes wallet debit after payment authorization
//   - payments.created: Validates funds availability for new payments
//
// The handler unmarshals the appropriate event type and delegates to the service layer.
func (h *WalletHandler) Handler(ctx context.Context, topic string, raw []byte) error {

	switch topic {
	case models.WalletDebitEventTopic:
		var event models.WalletDebitRequestedEvent

		if err := json.Unmarshal(raw, &event); err != nil {
			logrus.Errorf("Error unmarshalling WalletDebitRequestedEvent: %s", err.Error())
			return err
		}

		if err := h.WalletService.DebitBalance(ctx, event.UserID, event.Amount); err != nil {
			logrus.Errorf("Error validating funds: %s", err.Error())
			return err
		}

		logrus.Info("WalletDebitRequestedEvent handled successfully")
	case models.PaymentCreatedEventTopic:
		var event models.PaymentCreatedEvent

		if err := json.Unmarshal(raw, &event); err != nil {
			logrus.Errorf("Error unmarshalling PaymentCreatedEvent: %s", err.Error())
			return err
		}

		if err := h.WalletService.ValidateFunds(ctx, event); err != nil {
			logrus.Errorf("Error validating funds: %s", err.Error())
			return err
		}

		logrus.Info("PaymentCreatedEvent handled successfully")
	}

	return nil
}

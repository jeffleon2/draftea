package handler

import (
	"context"
	"encoding/json"

	"github.com/jeffleon2/draftea-wallet-service/internal/models"
	"github.com/sirupsen/logrus"
)

type WalletServiceIn interface {
	ValidateFunds(ctx context.Context, event models.PaymentCreatedEvent) error
	DebitBalance(ctx context.Context, userID string, amount float64) error
}

type WalletHandler struct {
	WalletService WalletServiceIn
}

func Wallet(s WalletServiceIn) *WalletHandler {
	return &WalletHandler{
		WalletService: s,
	}
}

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

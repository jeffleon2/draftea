package handler

import (
	"context"
	"encoding/json"

	"github.com/jeffleon2/draftea-fraud-service/internal/models"
	"github.com/sirupsen/logrus"
)

type FraudServiceIn interface {
	EvaluatePayment(context.Context, models.PaymentCreatedEvent) error
}

type FraudHandler struct {
	FraudService FraudServiceIn
}

func Fraud(s FraudServiceIn) *FraudHandler {
	return &FraudHandler{
		FraudService: s,
	}
}

func (h *FraudHandler) Handler(ctx context.Context, raw []byte) error {
	var event models.PaymentCreatedEvent

	if err := json.Unmarshal(raw, &event); err != nil {
		logrus.Errorf("Error unmarshalling PaymentCreatedEvent: %s", err.Error())
		return err
	}

	err := h.FraudService.EvaluatePayment(ctx, event)
	if err != nil {
		logrus.Errorf("Error evaluating fraud: %s", err.Error())
		return err
	}

	logrus.Info("PaymentCreatedEvent handled successfully")

	return nil
}

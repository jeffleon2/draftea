package handler

import (
	"context"
	"encoding/json"

	"github.com/jeffleon2/draftea-fraud-service/internal/models"
	"github.com/sirupsen/logrus"
)

// FraudServiceIn defines the interface for fraud evaluation operations.
// Implementations should handle fraud detection logic for payment transactions.
type FraudServiceIn interface {
	EvaluatePayment(context.Context, models.PaymentCreatedEvent) error
}

// FraudHandler processes incoming payment events for fraud detection.
// It acts as the entry point for the fraud service, unmarshalling events
// and delegating fraud evaluation to the FraudService.
type FraudHandler struct {
	FraudService FraudServiceIn
}

// Fraud creates a new FraudHandler with the provided fraud service implementation.
// The handler will use this service to evaluate payments for potential fraud.
func Fraud(s FraudServiceIn) *FraudHandler {
	return &FraudHandler{
		FraudService: s,
	}
}

// Handler processes raw payment event messages from Kafka.
// It unmarshals the JSON payload into a PaymentCreatedEvent and delegates
// fraud evaluation to the FraudService. Returns an error if unmarshalling
// fails or if fraud evaluation encounters an issue.
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

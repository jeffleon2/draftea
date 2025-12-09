package service

import (
	"context"
	"log"
	"time"

	"github.com/jeffleon2/draftea-fraud-service/internal/models"
	"github.com/sirupsen/logrus"
)

// Publisher defines the interface for publishing events to Kafka topics.
type Publisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
}

// FraudService implements fraud detection logic for payment transactions.
// It evaluates payments based on configurable rules and publishes the results
// to Kafka for downstream processing.
type FraudService struct {
	Publisher Publisher
}

// NewFraudService creates a new FraudService with the provided publisher.
// The publisher is used to send fraud check results to the payments.checked topic.
func NewFraudService(p Publisher) *FraudService {
	return &FraudService{
		Publisher: p,
	}
}

// EvaluatePayment analyzes a payment for potential fraud.
// Current implementation flags transactions over $10,000 as suspicious.
// Results are published to the payments.checked topic with either APPROVED or DECLINED status.
//
// The method includes a 30-second delay to simulate fraud analysis processing time.
// In production, this would be replaced with actual fraud detection algorithms.
func (s *FraudService) EvaluatePayment(ctx context.Context, event models.PaymentCreatedEvent) error {
	log.Println("Evaluating fraud for payment:", event.ID)
	var reason string
	var status = models.PaymentStatusApproved

	if event.Amount > 10000 {
		reason = "High-value transaction suspicious"
		status = models.PaymentStatusDeclined
		logrus.Errorf("High value payment detected - potential fraud %s", event.ID)
	}

	approved := models.FraudCheckEvent{
		ID:        event.ID,
		TraceID:   event.TraceID,
		CheckedAt: time.Now(),
		Reason:    reason,
		Status:    status,
	}

	time.Sleep(30 * time.Second)

	log.Println("✅ Fraud evaluation completed for:", event.ID)
	return s.Publisher.Publish(ctx, models.TopicPaymentChecked, approved)
}

package service

import (
	"context"
	"log"
	"time"

	"github.com/jeffleon2/draftea-fraud-service/internal/models"
	"github.com/sirupsen/logrus"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
}

type FraudService struct {
	Publisher Publisher
}

func NewFraudService(p Publisher) *FraudService {
	return &FraudService{
		Publisher: p,
	}
}

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

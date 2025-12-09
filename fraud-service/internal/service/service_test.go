package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jeffleon2/draftea-fraud-service/internal/models"
	"github.com/jeffleon2/draftea-fraud-service/internal/service"
	"github.com/jeffleon2/draftea-fraud-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEvaluatePayment_LowValuePayment_Approved(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	fraudService := service.NewFraudService(mockPublisher)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-123",
		Amount:     5000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-456",
		TraceID:    "trace-789",
	}

	mockPublisher.EXPECT().
		Publish(ctx, models.TopicPaymentChecked, mock.MatchedBy(func(evt models.FraudCheckEvent) bool {
			return evt.ID == event.ID &&
				evt.TraceID == event.TraceID &&
				evt.Status == models.PaymentStatusApproved &&
				evt.Reason == ""
		})).
		Return(nil).
		Once()

	err := fraudService.EvaluatePayment(ctx, event)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestEvaluatePayment_HighValuePayment_Declined(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	fraudService := service.NewFraudService(mockPublisher)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-999",
		Amount:     15000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-suspicious",
		TraceID:    "trace-suspicious",
	}

	mockPublisher.EXPECT().
		Publish(ctx, models.TopicPaymentChecked, mock.MatchedBy(func(evt models.FraudCheckEvent) bool {
			return evt.ID == event.ID &&
				evt.TraceID == event.TraceID &&
				evt.Status == models.PaymentStatusDeclined &&
				evt.Reason == "High-value transaction suspicious"
		})).
		Return(nil).
		Once()

	err := fraudService.EvaluatePayment(ctx, event)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestEvaluatePayment_ExactThreshold_Approved(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	fraudService := service.NewFraudService(mockPublisher)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-threshold",
		Amount:     10000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-edge",
		TraceID:    "trace-edge",
	}

	mockPublisher.EXPECT().
		Publish(ctx, models.TopicPaymentChecked, mock.MatchedBy(func(evt models.FraudCheckEvent) bool {
			return evt.ID == event.ID &&
				evt.Status == models.PaymentStatusApproved
		})).
		Return(nil).
		Once()

	err := fraudService.EvaluatePayment(ctx, event)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestEvaluatePayment_PublisherError(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	fraudService := service.NewFraudService(mockPublisher)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-error",
		Amount:     5000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-error",
		TraceID:    "trace-error",
	}

	expectedError := errors.New("kafka publish failed")

	mockPublisher.EXPECT().
		Publish(ctx, models.TopicPaymentChecked, mock.Anything).
		Return(expectedError).
		Once()

	err := fraudService.EvaluatePayment(ctx, event)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
}

func TestEvaluatePayment_ZeroAmount_Approved(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	fraudService := service.NewFraudService(mockPublisher)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-zero",
		Amount:     0.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-zero",
		TraceID:    "trace-zero",
	}

	mockPublisher.EXPECT().
		Publish(ctx, models.TopicPaymentChecked, mock.MatchedBy(func(evt models.FraudCheckEvent) bool {
			return evt.Status == models.PaymentStatusApproved
		})).
		Return(nil).
		Once()

	err := fraudService.EvaluatePayment(ctx, event)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestEvaluatePayment_NegativeAmount_Approved(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	fraudService := service.NewFraudService(mockPublisher)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-negative",
		Amount:     -100.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "refund",
		CustomerID: "customer-refund",
		TraceID:    "trace-refund",
	}

	mockPublisher.EXPECT().
		Publish(ctx, models.TopicPaymentChecked, mock.MatchedBy(func(evt models.FraudCheckEvent) bool {
			return evt.Status == models.PaymentStatusApproved
		})).
		Return(nil).
		Once()

	err := fraudService.EvaluatePayment(ctx, event)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestEvaluatePayment_CheckedAtTimestamp(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	fraudService := service.NewFraudService(mockPublisher)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-timestamp",
		Amount:     5000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-timestamp",
		TraceID:    "trace-timestamp",
	}

	beforeTime := time.Now()

	mockPublisher.EXPECT().
		Publish(ctx, models.TopicPaymentChecked, mock.MatchedBy(func(evt models.FraudCheckEvent) bool {
			return !evt.CheckedAt.IsZero() &&
				evt.CheckedAt.After(beforeTime) &&
				evt.CheckedAt.Before(time.Now().Add(35*time.Second))
		})).
		Return(nil).
		Once()

	err := fraudService.EvaluatePayment(ctx, event)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestNewFraudService(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)

	fraudService := service.NewFraudService(mockPublisher)

	assert.NotNil(t, fraudService)
	assert.Equal(t, mockPublisher, fraudService.Publisher)
}

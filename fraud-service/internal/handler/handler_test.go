package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jeffleon2/draftea-fraud-service/internal/handler"
	"github.com/jeffleon2/draftea-fraud-service/internal/handler/mocks"
	"github.com/jeffleon2/draftea-fraud-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Success(t *testing.T) {
	mockService := mocks.NewMockFraudServiceIn(t)
	h := handler.Fraud(mockService)

	event := models.PaymentCreatedEvent{
		ID:         "payment-123",
		Amount:     5000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-456",
		TraceID:    "trace-789",
	}

	eventBytes, err := json.Marshal(event)
	assert.NoError(t, err)

	ctx := context.Background()

	mockService.EXPECT().
		EvaluatePayment(ctx, event).
		Return(nil).
		Once()

	err = h.Handler(ctx, eventBytes)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestHandler_UnmarshalError(t *testing.T) {
	mockService := mocks.NewMockFraudServiceIn(t)
	h := handler.Fraud(mockService)

	invalidJSON := []byte(`{"invalid json`)
	ctx := context.Background()

	err := h.Handler(ctx, invalidJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected end of JSON input")
	mockService.AssertNotCalled(t, "EvaluatePayment", mock.Anything, mock.Anything)
}

func TestHandler_ServiceError(t *testing.T) {
	mockService := mocks.NewMockFraudServiceIn(t)
	h := handler.Fraud(mockService)

	event := models.PaymentCreatedEvent{
		ID:         "payment-123",
		Amount:     5000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-456",
		TraceID:    "trace-789",
	}

	eventBytes, err := json.Marshal(event)
	assert.NoError(t, err)

	ctx := context.Background()
	expectedError := errors.New("service evaluation failed")

	mockService.EXPECT().
		EvaluatePayment(ctx, event).
		Return(expectedError).
		Once()

	err = h.Handler(ctx, eventBytes)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockService.AssertExpectations(t)
}

func TestHandler_HighValuePayment(t *testing.T) {
	mockService := mocks.NewMockFraudServiceIn(t)
	h := handler.Fraud(mockService)

	event := models.PaymentCreatedEvent{
		ID:         "payment-999",
		Amount:     15000.0,
		Currency:   "USD",
		Status:     "PENDING",
		Method:     "credit_card",
		CustomerID: "customer-suspicious",
		TraceID:    "trace-suspicious",
	}

	eventBytes, err := json.Marshal(event)
	assert.NoError(t, err)

	ctx := context.Background()

	mockService.EXPECT().
		EvaluatePayment(ctx, event).
		Return(nil).
		Once()

	err = h.Handler(ctx, eventBytes)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestFraud_Constructor(t *testing.T) {
	mockService := mocks.NewMockFraudServiceIn(t)

	h := handler.Fraud(mockService)

	assert.NotNil(t, h)
	assert.Equal(t, mockService, h.FraudService)
}

func TestHandler_EmptyPayload(t *testing.T) {
	mockService := mocks.NewMockFraudServiceIn(t)
	h := handler.Fraud(mockService)

	emptyJSON := []byte(`{}`)
	ctx := context.Background()

	var emptyEvent models.PaymentCreatedEvent
	mockService.EXPECT().
		EvaluatePayment(ctx, emptyEvent).
		Return(nil).
		Once()

	err := h.Handler(ctx, emptyJSON)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

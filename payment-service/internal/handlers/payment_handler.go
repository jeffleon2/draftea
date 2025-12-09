package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeffleon2/draftea-payment-service/internal/models"
	"github.com/jeffleon2/draftea-payment-service/internal/models/dto"
	"github.com/sirupsen/logrus"
)

// PaymentService defines the interface for payment business logic operations.
type PaymentService interface {
	CreatePayment(ctx context.Context, payment *dto.Payment) error
	UpdatePaymentFlags(ctx context.Context, paymentID string, walletApproved, fraudClean *bool, failureReason string) error
}

// PaymentHandler handles HTTP requests and Kafka events for payment operations.
// It acts as the adapter layer between HTTP/Kafka and the payment service business logic.
type PaymentHandler struct {
	Service PaymentService
}

// NewPaymentHandler creates a new PaymentHandler with the provided service.
func NewPaymentHandler(s PaymentService) *PaymentHandler {
	return &PaymentHandler{Service: s}
}

// CreatePayment handles POST /payments HTTP requests.
// It validates the request body, delegates to the service layer,
// and returns 201 Created on success or appropriate error status.
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req dto.Payment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.Service.CreatePayment(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// HandleEvents processes Kafka events for payment verification updates.
// It handles two types of events:
//   - wallet.funds.verified: Updates wallet approval status
//   - payments.checked: Updates fraud check status
//
// The handler unmarshals the event, extracts the relevant status,
// and calls UpdatePaymentFlags to update the payment's verification state.
func (h *PaymentHandler) HandleEvents(ctx context.Context, topic string, value []byte) error {
	var walletStatus *bool
	var fraudStatus *bool
	var paymentID string
	var failureReason string

	switch topic {
	case models.WalletTopic2Subscribe:
		var event models.WalletResponseEvent
		if err := json.Unmarshal(value, &event); err != nil {
			logrus.Errorf("Error parsing Wallet response event %s", err.Error())
			return fmt.Errorf("error parsing Wallet response event %w", err)
		}
		flag := event.Status == models.PaymentStatusApproved
		walletStatus = &flag
		paymentID = event.PaymentID
		failureReason = event.Reason
	case models.FraudTopic2Subscribe:
		var event models.FraudCheckEvent
		if err := json.Unmarshal(value, &event); err != nil {
			logrus.Errorf("Error parsing Fraud check event %s", err.Error())
			return fmt.Errorf("error parsing Fraud check event %w", err)
		}
		flag := event.Status == models.PaymentStatusApproved
		fraudStatus = &flag
		paymentID = event.ID
		failureReason = event.Reason
	default:
		logrus.Errorf("topic not allowed %s", topic)
		return fmt.Errorf("topic not allowed %s", topic)
	}

	if err := h.Service.UpdatePaymentFlags(ctx, paymentID, walletStatus, fraudStatus, failureReason); err != nil {
		return fmt.Errorf("error updating payment flags %w", err)
	}

	return nil
}

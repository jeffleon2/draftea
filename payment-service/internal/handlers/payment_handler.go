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

type PaymentService interface {
	CreatePayment(ctx context.Context, payment *dto.Payment) error
	UpdatePaymentFlags(ctx context.Context, paymentID string, walletApproved, fraudClean *bool, failureReason string) error
}

type PaymentHandler struct {
	Service PaymentService
}

func NewPaymentHandler(s PaymentService) *PaymentHandler {
	return &PaymentHandler{Service: s}
}

// POST /payments
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

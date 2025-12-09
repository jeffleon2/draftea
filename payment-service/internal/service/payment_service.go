package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/jeffleon2/draftea-payment-service/internal/models"
	"github.com/jeffleon2/draftea-payment-service/internal/models/dto"
)

var paymentLocks = make(map[string]*sync.Mutex)
var mu sync.Mutex

// PaymentRepo defines the interface for payment data persistence operations.
// Implementations should handle CRUD operations for payment entities.
type PaymentRepo interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id string) (*models.Payment, error)
	GetAll(ctx context.Context) (*[]models.Payment, error)
	Update(ctx context.Context, payment *models.Payment, id string) error
	Delete(ctx context.Context, id string) error
}

// Publisher defines the interface for publishing events to Kafka topics.
type Publisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
}

// PaymentService orchestrates payment processing workflows.
// It coordinates between fraud detection, wallet verification, and payment authorization,
// implementing a saga pattern with parallel verification of fraud and funds.
type PaymentService struct {
	Repo      PaymentRepo
	Publisher Publisher
}

// NewPaymentService creates a new PaymentService with the provided repository and publisher.
// The service uses the repository for payment persistence and the publisher for event-driven communication.
func NewPaymentService(repo PaymentRepo, publisher Publisher) *PaymentService {
	return &PaymentService{
		Repo:      repo,
		Publisher: publisher,
	}
}

// CreatePayment initiates a new payment transaction.
// It validates the payment data, persists it to the database with PENDING status,
// and publishes a payments.created event to trigger parallel fraud and funds verification.
//
// The payment starts with both fraud_checked and funds_verified flags set to false.
// These flags will be updated by the fraud and wallet services asynchronously.
func (s *PaymentService) CreatePayment(ctx context.Context, paymentDTO *dto.Payment) error {
	paymentDTO.Sanitize()
	payment := paymentDTO.ToEntity()
	if err := payment.Validate(); err != nil {
		return err
	}

	if err := s.Repo.Create(ctx, payment); err != nil {
		return err
	}

	event := models.PaymentCreatedEvent{
		ID:         payment.ID,
		Amount:     payment.Amount,
		Currency:   string(payment.Currency),
		Status:     string(payment.Status),
		Method:     string(payment.Method),
		CustomerID: payment.CustomerID,
		TraceID:    payment.TraceID,
		CreatedAt:  payment.CreatedAt,
	}

	return s.Publisher.Publish(ctx, models.PaymentCreatedEventTopic, event)
}

// UpdatePaymentFlags updates the verification status flags for a payment.
// This method is called by event handlers when receiving fraud check or wallet verification results.
//
// Parameters:
//   - walletApproved: pointer to bool indicating wallet funds verification result (nil if not updating)
//   - fraudClean: pointer to bool indicating fraud check result (nil if not updating)
//   - failureReason: reason for failure if either check declined the payment
//
// If either check fails, the payment status is immediately set to FAILED.
// If both checks pass, CompletePaymentIfReady is called to authorize the payment and trigger wallet debit.
func (s *PaymentService) UpdatePaymentFlags(
	ctx context.Context,
	paymentID string,
	walletApproved *bool,
	fraudClean *bool,
	failureReason string,
) error {

	payment, err := s.Repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	fmt.Println("Payment ", payment)

	if walletApproved != nil {
		if !*walletApproved {
			payment.FailedReason = failureReason
			payment.Status = models.StatusFailed
		}
		payment.WalletApproved = *walletApproved
	}
	if fraudClean != nil {
		if !*fraudClean {
			payment.FailedReason = failureReason
			payment.Status = models.StatusFailed
		}
		payment.FraudCleared = *fraudClean
	}

	if err := s.Repo.Update(ctx, payment, paymentID); err != nil {
		return err
	}

	fmt.Println("Update successfully flags")

	return s.CompletePaymentIfReady(ctx, payment)
}

// CompletePaymentIfReady checks if a payment is ready for authorization.
// A payment is ready when both fraud_checked and funds_verified flags are true.
//
// This method uses a mutex lock per payment ID to prevent race conditions when
// multiple verification events arrive concurrently. If both verifications pass,
// the payment status is updated to AUTHORIZED and a wallet.debit.requested event
// is published to trigger the actual wallet debit operation.
func (s *PaymentService) CompletePaymentIfReady(ctx context.Context, payment *models.Payment) error {
	lock := getLock(payment.ID)
	lock.Lock()
	defer lock.Unlock()

	if payment.Status == models.StatusAuthorized || payment.Status == models.StatusFailed || !payment.WalletApproved || !payment.FraudCleared {
		return nil
	}

	if payment.WalletApproved && payment.FraudCleared {
		payment.Status = models.StatusAuthorized
		payment.FailedReason = ""
	}

	if err := s.Repo.Update(ctx, payment, payment.ID); err != nil {
		return err
	}

	event := models.WalletDebitRequestedEvent{
		PaymentID: payment.ID,
		UserID:    payment.CustomerID,
		Amount:    payment.Amount,
		Reason:    "PAYMENT_COMPLETE",
		TraceID:   payment.TraceID,
	}

	return s.Publisher.Publish(ctx, models.WalletDebitEventTopic, event)
}

func getLock(paymentID string) *sync.Mutex {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := paymentLocks[paymentID]; !ok {
		paymentLocks[paymentID] = &sync.Mutex{}
	}
	return paymentLocks[paymentID]
}

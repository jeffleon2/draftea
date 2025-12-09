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

type PaymentRepo interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id string) (*models.Payment, error)
	GetAll(ctx context.Context) (*[]models.Payment, error)
	Update(ctx context.Context, payment *models.Payment, id string) error
	Delete(ctx context.Context, id string) error
}

type Publisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
}

type PaymentService struct {
	Repo      PaymentRepo
	Publisher Publisher
}

func NewPaymentService(repo PaymentRepo, publisher Publisher) *PaymentService {
	return &PaymentService{
		Repo:      repo,
		Publisher: publisher,
	}
}

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

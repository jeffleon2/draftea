package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeffleon2/draftea-wallet-service/internal/models"
)

// WalletRepo defines the interface for wallet data persistence operations.
// Implementations should handle CRUD operations for wallet entities.
type WalletRepo interface {
	Create(ctx context.Context, payment *models.Wallet) error
	GetByID(ctx context.Context, id string) (*models.Wallet, error)
	GetBy(ctx context.Context, key string, value interface{}) (*[]models.Wallet, error)
	GetAll(ctx context.Context) (*[]models.Wallet, error)
	Update(ctx context.Context, wallet *models.Wallet, id string) error
	Delete(ctx context.Context, id string) error
}

// Publisher defines the interface for publishing events to Kafka topics.
type Publisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
}

// WalletService manages wallet operations including funds verification and balance debits.
// It participates in the payment processing saga by verifying available funds
// and executing debits when authorized by the payment service.
type WalletService struct {
	Publisher  Publisher
	WalletRepo WalletRepo
}

// NewWalletService creates a new WalletService with the provided publisher and repository.
// The publisher is used for event-driven communication and the repository for wallet persistence.
func NewWalletService(p Publisher, w WalletRepo) *WalletService {
	return &WalletService{
		Publisher:  p,
		WalletRepo: w,
	}
}

// ValidateFunds checks if a user's wallet has sufficient balance for a payment.
// This method is called when a payments.created event is received.
//
// It verifies the wallet balance against the requested payment amount and publishes
// a wallet.funds.verified event with either APPROVED or DECLINED status.
// This is a verification-only operation; no funds are debited at this stage.
//
// Returns an error if the wallet is not found or if there's a database/publishing error.
func (s *WalletService) ValidateFunds(ctx context.Context, event models.PaymentCreatedEvent) error {
	var reasonDeclained string
	status := models.WalletStatusApproved

	wallet, err := s.WalletRepo.GetBy(ctx, "user_id", event.CustomerID)
	if err != nil {
		return err
	}

	if wallet == nil {
		return errors.New("wallet not found")
	}

	firstWallet := (*wallet)[0]
	if firstWallet.Balance < event.Amount {
		status = models.WalletStatusDeclined
		reasonDeclained = "Insufficient funds"
	}

	walletResponse := models.WalletResponseEvent{
		PaymentID: event.ID,
		UserID:    event.CustomerID,
		Amount:    event.Amount,
		Status:    status,
		Reason:    reasonDeclained,
	}

	return s.Publisher.Publish(ctx, models.WalletResponseTopic, walletResponse)
}

// DebitBalance deducts the specified amount from a user's wallet.
// This method is called when a wallet.debit.requested event is received,
// which only happens after both fraud and funds verifications have passed.
//
// The method retrieves the user's wallet, subtracts the amount from the balance,
// and persists the updated wallet. It returns an error if the wallet is not found
// or if the database update fails.
//
// Note: This method assumes funds were already verified by ValidateFunds.
// The actual balance check happened earlier in the payment flow.
func (s *WalletService) DebitBalance(ctx context.Context, userID string, amount float64) error {
	fmt.Println("Amount to discound", userID, amount)
	wallet, err := s.WalletRepo.GetBy(ctx, "user_id", userID)
	if err != nil {
		return err
	}

	if wallet == nil || len(*wallet) == 0 {
		return errors.New("wallet not found")
	}

	firstWallet := (*wallet)[0]
	firstWallet.Balance -= amount

	return s.WalletRepo.Update(ctx, &firstWallet, firstWallet.ID)
}

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeffleon2/draftea-wallet-service/internal/models"
)

type WalletRepo interface {
	Create(ctx context.Context, payment *models.Wallet) error
	GetByID(ctx context.Context, id string) (*models.Wallet, error)
	GetBy(ctx context.Context, key string, value interface{}) (*[]models.Wallet, error)
	GetAll(ctx context.Context) (*[]models.Wallet, error)
	Update(ctx context.Context, wallet *models.Wallet, id string) error
	Delete(ctx context.Context, id string) error
}

type Publisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
}

type WalletService struct {
	Publisher  Publisher
	WalletRepo WalletRepo
}

func NewWalletService(p Publisher, w WalletRepo) *WalletService {
	return &WalletService{
		Publisher:  p,
		WalletRepo: w,
	}
}

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

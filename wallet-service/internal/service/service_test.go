package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffleon2/draftea-wallet-service/internal/models"
	"github.com/jeffleon2/draftea-wallet-service/internal/service"
	"github.com/jeffleon2/draftea-wallet-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateFunds_SufficientBalance(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)
	walletService := service.NewWalletService(mockPublisher, mockRepo)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-123",
		Amount:     50.0,
		Currency:   "USD",
		CustomerID: "customer-456",
	}

	wallets := &[]models.Wallet{
		{
			ID:      "wallet-1",
			UserID:  "customer-456",
			Balance: 100.0,
		},
	}

	mockRepo.EXPECT().
		GetBy(ctx, "user_id", event.CustomerID).
		Return(wallets, nil).
		Once()

	mockPublisher.EXPECT().
		Publish(ctx, models.WalletResponseTopic, mock.MatchedBy(func(evt models.WalletResponseEvent) bool {
			return evt.PaymentID == event.ID &&
				evt.UserID == event.CustomerID &&
				evt.Status == models.WalletStatusApproved &&
				evt.Amount == event.Amount &&
				evt.Reason == ""
		})).
		Return(nil).
		Once()

	err := walletService.ValidateFunds(ctx, event)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestValidateFunds_InsufficientBalance(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)
	walletService := service.NewWalletService(mockPublisher, mockRepo)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-789",
		Amount:     150.0,
		Currency:   "USD",
		CustomerID: "customer-poor",
	}

	wallets := &[]models.Wallet{
		{
			ID:      "wallet-2",
			UserID:  "customer-poor",
			Balance: 50.0,
		},
	}

	mockRepo.EXPECT().
		GetBy(ctx, "user_id", event.CustomerID).
		Return(wallets, nil).
		Once()

	mockPublisher.EXPECT().
		Publish(ctx, models.WalletResponseTopic, mock.MatchedBy(func(evt models.WalletResponseEvent) bool {
			return evt.PaymentID == event.ID &&
				evt.UserID == event.CustomerID &&
				evt.Status == models.WalletStatusDeclined &&
				evt.Amount == event.Amount &&
				evt.Reason == "Insufficient funds"
		})).
		Return(nil).
		Once()

	err := walletService.ValidateFunds(ctx, event)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestValidateFunds_WalletNotFound(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)
	walletService := service.NewWalletService(mockPublisher, mockRepo)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-404",
		Amount:     100.0,
		Currency:   "USD",
		CustomerID: "customer-nonexistent",
	}

	mockRepo.EXPECT().
		GetBy(ctx, "user_id", event.CustomerID).
		Return(nil, nil).
		Once()

	err := walletService.ValidateFunds(ctx, event)

	assert.Error(t, err)
	assert.Equal(t, "wallet not found", err.Error())
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestValidateFunds_RepoError(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)
	walletService := service.NewWalletService(mockPublisher, mockRepo)

	ctx := context.Background()
	event := models.PaymentCreatedEvent{
		ID:         "payment-error",
		Amount:     100.0,
		Currency:   "USD",
		CustomerID: "customer-error",
	}

	expectedError := errors.New("database connection failed")

	mockRepo.EXPECT().
		GetBy(ctx, "user_id", event.CustomerID).
		Return(nil, expectedError).
		Once()

	err := walletService.ValidateFunds(ctx, event)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestDebitBalance_Success(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)
	walletService := service.NewWalletService(mockPublisher, mockRepo)

	ctx := context.Background()
	userID := "user-123"
	amount := 50.0

	wallets := &[]models.Wallet{
		{
			ID:      "wallet-1",
			UserID:  userID,
			Balance: 100.0,
		},
	}

	mockRepo.EXPECT().
		GetBy(ctx, "user_id", userID).
		Return(wallets, nil).
		Once()

	mockRepo.EXPECT().
		Update(ctx, mock.MatchedBy(func(w *models.Wallet) bool {
			return w.Balance == 50.0
		}), "wallet-1").
		Return(nil).
		Once()

	err := walletService.DebitBalance(ctx, userID, amount)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDebitBalance_WalletNotFound(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)
	walletService := service.NewWalletService(mockPublisher, mockRepo)

	ctx := context.Background()
	userID := "user-nonexistent"
	amount := 50.0

	emptyWallets := &[]models.Wallet{}

	mockRepo.EXPECT().
		GetBy(ctx, "user_id", userID).
		Return(emptyWallets, nil).
		Once()

	err := walletService.DebitBalance(ctx, userID, amount)

	assert.Error(t, err)
	assert.Equal(t, "wallet not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestDebitBalance_RepoError(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)
	walletService := service.NewWalletService(mockPublisher, mockRepo)

	ctx := context.Background()
	userID := "user-error"
	amount := 50.0

	expectedError := errors.New("database error")

	mockRepo.EXPECT().
		GetBy(ctx, "user_id", userID).
		Return(nil, expectedError).
		Once()

	err := walletService.DebitBalance(ctx, userID, amount)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestNewWalletService(t *testing.T) {
	mockPublisher := mocks.NewMockPublisher(t)
	mockRepo := mocks.NewMockWalletRepo(t)

	walletService := service.NewWalletService(mockPublisher, mockRepo)

	assert.NotNil(t, walletService)
	assert.Equal(t, mockPublisher, walletService.Publisher)
	assert.Equal(t, mockRepo, walletService.WalletRepo)
}

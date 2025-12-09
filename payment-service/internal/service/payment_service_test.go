package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jeffleon2/draftea-payment-service/internal/models"
	"github.com/jeffleon2/draftea-payment-service/internal/models/dto"
	"github.com/jeffleon2/draftea-payment-service/internal/service"
	"github.com/jeffleon2/draftea-payment-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePayment_Success(t *testing.T) {
	mockRepo := mocks.NewMockPaymentRepo(t)
	mockPublisher := mocks.NewMockPublisher(t)
	paymentService := service.NewPaymentService(mockRepo, mockPublisher)

	ctx := context.Background()
	paymentDTO := &dto.Payment{
		Amount:     100.50,
		Currency:   "USD",
		Method:     "CREDIT_CARD",
		CustomerID: "customer-123",
	}

	mockRepo.EXPECT().
		Create(ctx, mock.AnythingOfType("*models.Payment")).
		Return(nil).
		Once()

	mockPublisher.EXPECT().
		Publish(ctx, models.PaymentCreatedEventTopic, mock.AnythingOfType("models.PaymentCreatedEvent")).
		Return(nil).
		Once()

	err := paymentService.CreatePayment(ctx, paymentDTO)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreatePayment_RepoError(t *testing.T) {
	mockRepo := mocks.NewMockPaymentRepo(t)
	mockPublisher := mocks.NewMockPublisher(t)
	paymentService := service.NewPaymentService(mockRepo, mockPublisher)

	ctx := context.Background()
	paymentDTO := &dto.Payment{
		Amount:     100.50,
		Currency:   "USD",
		Method:     "CREDIT_CARD",
		CustomerID: "customer-123",
	}

	expectedError := errors.New("database error")

	mockRepo.EXPECT().
		Create(ctx, mock.AnythingOfType("*models.Payment")).
		Return(expectedError).
		Once()

	err := paymentService.CreatePayment(ctx, paymentDTO)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreatePayment_PublisherError(t *testing.T) {
	mockRepo := mocks.NewMockPaymentRepo(t)
	mockPublisher := mocks.NewMockPublisher(t)
	paymentService := service.NewPaymentService(mockRepo, mockPublisher)

	ctx := context.Background()
	paymentDTO := &dto.Payment{
		Amount:     100.50,
		Currency:   "USD",
		Method:     "CREDIT_CARD",
		CustomerID: "customer-123",
	}

	expectedError := errors.New("kafka publish error")

	mockRepo.EXPECT().
		Create(ctx, mock.AnythingOfType("*models.Payment")).
		Return(nil).
		Once()

	mockPublisher.EXPECT().
		Publish(ctx, models.PaymentCreatedEventTopic, mock.AnythingOfType("models.PaymentCreatedEvent")).
		Return(expectedError).
		Once()

	err := paymentService.CreatePayment(ctx, paymentDTO)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestUpdatePaymentFlags_FraudDeclined(t *testing.T) {
	mockRepo := mocks.NewMockPaymentRepo(t)
	mockPublisher := mocks.NewMockPublisher(t)
	paymentService := service.NewPaymentService(mockRepo, mockPublisher)

	ctx := context.Background()
	paymentID := "payment-123"
	fraudClean := false
	failureReason := "fraud detected"

	existingPayment := &models.Payment{
		ID:             paymentID,
		Amount:         100.0,
		Status:         models.StatusPending,
		WalletApproved: false,
		FraudCleared:   false,
	}

	mockRepo.EXPECT().
		GetByID(ctx, paymentID).
		Return(existingPayment, nil).
		Once()

	mockRepo.EXPECT().
		Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.FraudCleared == false &&
				p.Status == models.StatusFailed &&
				p.FailedReason == failureReason
		}), paymentID).
		Return(nil).
		Once()

	err := paymentService.UpdatePaymentFlags(ctx, paymentID, nil, &fraudClean, failureReason)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePaymentFlags_WalletDeclined(t *testing.T) {
	mockRepo := mocks.NewMockPaymentRepo(t)
	mockPublisher := mocks.NewMockPublisher(t)
	paymentService := service.NewPaymentService(mockRepo, mockPublisher)

	ctx := context.Background()
	paymentID := "payment-456"
	walletApproved := false
	failureReason := "insufficient funds"

	existingPayment := &models.Payment{
		ID:             paymentID,
		Amount:         100.0,
		Status:         models.StatusPending,
		WalletApproved: false,
		FraudCleared:   false,
	}

	mockRepo.EXPECT().
		GetByID(ctx, paymentID).
		Return(existingPayment, nil).
		Once()

	mockRepo.EXPECT().
		Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.WalletApproved == false &&
				p.Status == models.StatusFailed &&
				p.FailedReason == failureReason
		}), paymentID).
		Return(nil).
		Once()
	err := paymentService.UpdatePaymentFlags(ctx, paymentID, &walletApproved, nil, failureReason)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePaymentFlags_BothApproved_TriggersDebit(t *testing.T) {
	mockRepo := mocks.NewMockPaymentRepo(t)
	mockPublisher := mocks.NewMockPublisher(t)
	paymentService := service.NewPaymentService(mockRepo, mockPublisher)

	ctx := context.Background()
	paymentID := "payment-789"
	walletApproved := true
	fraudClean := true

	existingPayment := &models.Payment{
		ID:             paymentID,
		Amount:         100.0,
		CustomerID:     "customer-123",
		TraceID:        "trace-123",
		Status:         models.StatusPending,
		WalletApproved: false,
		FraudCleared:   false,
	}

	mockRepo.EXPECT().
		GetByID(ctx, paymentID).
		Return(existingPayment, nil).
		Once()

	mockRepo.EXPECT().
		Update(ctx, mock.AnythingOfType("*models.Payment"), paymentID).
		Return(nil).
		Once()

	mockRepo.EXPECT().
		Update(ctx, mock.MatchedBy(func(p *models.Payment) bool {
			return p.Status == models.StatusAuthorized &&
				p.WalletApproved == true &&
				p.FraudCleared == true
		}), paymentID).
		Return(nil).
		Once()

	mockPublisher.EXPECT().
		Publish(ctx, models.WalletDebitEventTopic, mock.AnythingOfType("models.WalletDebitRequestedEvent")).
		Return(nil).
		Once()

	err := paymentService.UpdatePaymentFlags(ctx, paymentID, &walletApproved, &fraudClean, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestNewPaymentService(t *testing.T) {
	mockRepo := mocks.NewMockPaymentRepo(t)
	mockPublisher := mocks.NewMockPublisher(t)

	paymentService := service.NewPaymentService(mockRepo, mockPublisher)

	assert.NotNil(t, paymentService)
	assert.Equal(t, mockRepo, paymentService.Repo)
	assert.Equal(t, mockPublisher, paymentService.Publisher)
}

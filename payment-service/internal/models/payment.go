package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string
type Currency string
type PaymentMethod string

const (
	StatusPending    PaymentStatus = "PENDING"
	StatusAuthorized PaymentStatus = "AUTHORIZED"
	StatusFailed     PaymentStatus = "FAILED"
	StatusCancelled  PaymentStatus = "CANCELLED"

	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyMXN Currency = "MXN"
	CurrencyCOP Currency = "COP"

	MethodCreditCard   PaymentMethod = "CREDIT_CARD"
	MethodDebitCard    PaymentMethod = "DEBIT_CARD"
	MethodPaypal       PaymentMethod = "PAYPAL"
	MethodBankTransfer PaymentMethod = "BANK_TRANSFER"
)

type Payment struct {
	ID             string        `json:"id"`
	Amount         float64       `json:"amount"`
	Currency       Currency      `json:"currency"`
	Status         PaymentStatus `json:"status"`
	Method         PaymentMethod `json:"method"`
	CustomerID     string        `json:"customer_id"`
	WalletApproved bool
	FraudCleared   bool
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AuthorizedAt   time.Time `json:"authorized_at,omitempty"`
	FailedReason   string    `json:"failed_reason,omitempty"`
	TraceID        string    `json:"trace_id"`
}

func (p *Payment) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	return
}

func (p *Payment) Validate() error {
	if !p.Method.IsValid() {
		return fmt.Errorf("invalid payment method: %s", p.Method)
	}
	if !p.Currency.IsValid() {
		return fmt.Errorf("invalid currency: %s", p.Currency)
	}
	if p.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if p.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}

	return nil
}

func (m PaymentMethod) IsValid() bool {
	switch m {
	case MethodCreditCard, MethodDebitCard, MethodPaypal, MethodBankTransfer:
		return true
	default:
		return false
	}
}

func (c Currency) IsValid() bool {
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyMXN, CurrencyCOP:
		return true
	default:
		return false
	}
}

func (s PaymentStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusAuthorized, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

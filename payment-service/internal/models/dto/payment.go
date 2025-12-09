package dto

import (
	"strings"

	"github.com/jeffleon2/draftea-payment-service/internal/models"
)

type Payment struct {
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
	Method     string  `json:"method"`
	CustomerID string  `json:"customer_id"`
}

func (p *Payment) Sanitize() {
	p.Currency = strings.TrimSpace(p.Currency)
	p.Status = strings.TrimSpace(p.Status)
	p.Method = strings.TrimSpace(p.Method)
	p.CustomerID = strings.TrimSpace(p.CustomerID)

	p.Currency = strings.ToUpper(p.Currency)
	p.Status = strings.ToUpper(p.Status)
	p.Method = strings.ToUpper(p.Method)
}

func (p *Payment) ToEntity() *models.Payment {
	return &models.Payment{
		Amount:     p.Amount,
		Currency:   models.Currency(p.Currency),
		Method:     models.PaymentMethod(p.Method),
		CustomerID: p.CustomerID,
		Status:     models.StatusPending,
	}
}

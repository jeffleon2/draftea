package app

import handlers "github.com/jeffleon2/draftea-payment-service/internal/handlers"

func (a *App) RegisterRoutes(h *handlers.PaymentHandler) {
	app := a.Router.Group("/payments")
	app.POST("", h.CreatePayment)
}

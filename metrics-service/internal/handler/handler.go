package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jeffleon2/draftea-metric-service/internal/metrics"
	"github.com/jeffleon2/draftea-metric-service/internal/models"
)

type MetricsHandler struct {
}

func NewMetricHandler() *MetricsHandler {
	return &MetricsHandler{}
}

func (h *MetricsHandler) HandleEvents(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case models.TopicPaymentsCreated:
		var evt models.PaymentCreatedEvent
		if err := json.Unmarshal(value, &evt); err != nil {
			log.Println("Error unmarshaling PaymentCreatedEvent:", err)
			return err
		}

		metrics.PaymentsTotal.WithLabelValues("created").Inc()
		metrics.PaymentAmounts.WithLabelValues(evt.Currency).Observe(evt.Amount)

	case models.TopicPaymentsChecked:
		var evt models.FraudCheckEvent
		if err := json.Unmarshal(value, &evt); err != nil {
			log.Println("Error unmarshaling FraudCheckEvent:", err)
			return err
		}

		statusLabel := evt.Status
		if statusLabel == "" {
			statusLabel = "unknown"
		}

		metrics.FraudChecksTotal.WithLabelValues(statusLabel).Inc()

	case models.TopicWalletResponse:
		var evt models.WalletResponseEvent
		if err := json.Unmarshal(value, &evt); err != nil {
			log.Println("Error unmarshaling WalletResponseEvent:", err)
			return err
		}

		metrics.WalletResponsesTotal.WithLabelValues(evt.Status).Inc()
		metrics.WalletAmounts.WithLabelValues(evt.UserID).Observe(evt.Amount)

	case models.TopicWalletDebitRequested:
		var evt models.WalletDebitRequestedEvent
		if err := json.Unmarshal(value, &evt); err != nil {
			log.Println("Error unmarshaling WalletDebitRequestedEvent:", err)
			return err
		}

		metrics.WalletDebits.WithLabelValues(evt.UserID).Observe(evt.Amount)

	default:
		log.Println("Evento desconocido:", topic)
	}

	return nil
}

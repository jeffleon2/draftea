package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	PaymentsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_total",
			Help: "Número total de pagos procesados",
		},
		[]string{"status"},
	)

	PaymentAmounts = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_amounts",
			Help:    "Distribución de montos de pagos",
			Buckets: prometheus.LinearBuckets(0, 50, 20),
		},
		[]string{"currency"},
	)

	FraudChecksTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fraud_checks_total",
			Help: "Número total de cheques de fraude",
		},
		[]string{"status"},
	)

	WalletResponsesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wallet_responses_total",
			Help: "Número total de respuestas de wallet",
		},
		[]string{"status"},
	)

	WalletAmounts = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wallet_amounts",
			Help:    "Montos observados en wallet",
			Buckets: prometheus.LinearBuckets(0, 50, 20),
		},
		[]string{"user_id"},
	)

	WalletDebits = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wallet_debits",
			Help:    "Distribución de débitos de wallet",
			Buckets: prometheus.LinearBuckets(0, 50, 20),
		},
		[]string{"user_id"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(
		PaymentsTotal,
		PaymentAmounts,
		FraudChecksTotal,
		WalletResponsesTotal,
		WalletAmounts,
		WalletDebits,
	)
}

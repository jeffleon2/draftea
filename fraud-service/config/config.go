package config

import (
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func New() (*Config, error) {
	var Config Config
	err := godotenv.Load(".env")
	if err != nil {
		logrus.Error("Error can't get the environment variables by file")
	}
	if err := env.Parse(&Config); err != nil {
		logrus.Fatalf("Error initializing: %s", err.Error())
		os.Exit(1)
	}
	return &Config, nil
}

type Config struct {
	APP
	Kafka
}

type APP struct {
	PORT string `env:"APP_PORT" envDefault:"8090"`
}

type Kafka struct {
	Brokers              string        `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	FraudConsumerGroup   string        `env:"KAFKA_FRAUD_GROUP_ID"   envDefault:"fraud-service"`
	PaymentConsumerGroup string        `env:"KAFKA_PAYMENT_GROUP_ID" envDefault:"payment-service"`
	SubscriberTopics     string        `env:"KAFKA_SUBSCRIBER_TOPICS" envDefault:"payments.created"`
	PublishTopics        string        `env:"KAFKA_PUBLISH_TOPICS" envDefault:"payments.checked,payments.dlq"`
	RetryMaxAttempts     int           `env:"KAFKA_RETRY_MAX_ATTEMPTS" envDefault:"5"`
	RetryBaseDelay       time.Duration `env:"KAFKA_RETRY_BASE_DELAY" envDefault:"100ms"`
	RetryMaxDelay        time.Duration `env:"KAFKA_RETRY_MAX_DELAY" envDefault:"10s"`
	RetryJitter          bool          `env:"KAFKA_RETRY_JITTER" envDefault:"true"`
}

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Jitter      bool
}

func (k Kafka) GetRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: k.RetryMaxAttempts,
		BaseDelay:   k.RetryBaseDelay,
		MaxDelay:    k.RetryMaxDelay,
		Jitter:      k.RetryJitter,
	}
}

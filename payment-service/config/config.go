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
	DB
	Kafka
}

type DB struct {
	HOST     string `env:"DB_HOST"`
	USER     string `env:"DB_USER"`
	PASSWORD string `env:"DB_PASSWORD"`
	NAME     string `env:"DB_NAME"`
	PORT     string `env:"DB_PORT"`
	SSLMODE  string `env:"DB_SSLMODE"`
}

type APP struct {
	PORT string `env:"APP_PORT" envDefault:"8080"`
}

type Kafka struct {
	Brokers              string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	PaymentConsumerGroup string `env:"KAFKA_PAYMENT_GROUP_ID" envDefault:"payment-service"`
	PublishTopics        string `env:"KAFKA_PUBLISH_TOPICS" envDefault:"payments.created,wallet.debit.requested,payments.dlq"`
	SubscriberTopics     string `env:"KAFKA_SUBSCRIBER_TOPICS" envDefault:"payments.checked,wallet.funds.verified"`

	RetryMaxAttempts int           `env:"KAFKA_RETRY_MAX_ATTEMPTS" envDefault:"5"`
	RetryBaseDelay   time.Duration `env:"KAFKA_RETRY_BASE_DELAY" envDefault:"100ms"`
	RetryMaxDelay    time.Duration `env:"KAFKA_RETRY_MAX_DELAY" envDefault:"10s"`
	RetryJitter      bool          `env:"KAFKA_RETRY_JITTER" envDefault:"true"`
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

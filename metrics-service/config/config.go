package config

import (
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func New() (*Config, error) {
	var Config Config
	if os.Getenv("GO_ENV") == "local" {
		_ = godotenv.Load(".env")
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
	Brokers          string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	ConsumerGroup    string `env:"KAFKA_SUBSCRIBER_GROUP_ID" envDefault:"metric-service"`
	SubscriberTopics string `env:"KAFKA_SUBSCRIBER_TOPICS" envDefault:"payments.created,payments.checked,wallet.response,wallet.debit.requested"`
}

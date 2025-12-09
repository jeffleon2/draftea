package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jeffleon2/draftea-wallet-service/config"
	"github.com/jeffleon2/draftea-wallet-service/internal/database"
	"github.com/jeffleon2/draftea-wallet-service/internal/handler"
	"github.com/jeffleon2/draftea-wallet-service/internal/models"
	"github.com/jeffleon2/draftea-wallet-service/internal/publisher"
	"github.com/jeffleon2/draftea-wallet-service/internal/repository/posgrest"
	"github.com/jeffleon2/draftea-wallet-service/internal/service"
	"github.com/jeffleon2/draftea-wallet-service/internal/subscriber"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		fmt.Println("Error reading config file", err)
		os.Exit(1)
	}

	brokers := strings.Split(cfg.Kafka.Brokers, ",")
	publishTopics := strings.Split(cfg.Kafka.PublishTopics, ",")
	subscriberTopics := strings.Split(cfg.Kafka.SubscriberTopics, ",")

	db, err := cfg.DB.GormConnect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Wallet{}); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	if os.Getenv("GO_ENV") == "local" {
		if err := database.SeedWallets(db); err != nil {
			log.Printf("Warning: failed to seed wallets: %v", err)
		}
	}

	// TODO leon: arreglar brokers permitir []string
	publishers := publisher.NewKafkaPublisher(brokers[0], publishTopics, cfg.Kafka.GetRetryConfig())

	multiConsumer := subscriber.NewMultiTopicConsumer(brokers, subscriberTopics, cfg.Kafka.WalletConsumerGroup, publishers, cfg.Kafka.GetRetryConfig())
	walletRepo := posgrest.New[models.Wallet](db)
	walletService := service.NewWalletService(publishers, walletRepo)
	walletHandler := handler.Wallet(walletService)

	multiConsumer.Listen(ctx, func(topic string, value []byte) error {
		log.Printf("📩 Received event → topic=%s value=%s\n", topic, string(value))
		walletHandler.Handler(ctx, topic, value)
		return nil
	})

	<-ctx.Done()

	for _, reader := range multiConsumer.Readers {
		if err := reader.Close(); err != nil {
			log.Println("Error closing consumer:", err)
		}
	}

	log.Println("Wallet service stopped")
}

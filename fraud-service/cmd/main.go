package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jeffleon2/draftea-fraud-service/config"
	"github.com/jeffleon2/draftea-fraud-service/internal/handler"
	"github.com/jeffleon2/draftea-fraud-service/internal/publisher"
	"github.com/jeffleon2/draftea-fraud-service/internal/service"
	"github.com/jeffleon2/draftea-fraud-service/internal/subscriber"
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
	subscriberTopic := strings.Split(cfg.Kafka.SubscriberTopics, ",")
	publishers := publisher.NewKafkaPublisher(brokers[0], publishTopics, cfg.Kafka.GetRetryConfig())
	multiConsumer := subscriber.NewMultiTopicConsumer(brokers, subscriberTopic, cfg.Kafka.PaymentConsumerGroup, publishers, cfg.Kafka.GetRetryConfig())
	fraudService := service.NewFraudService(publishers)
	FraudHandler := handler.Fraud(fraudService)

	multiConsumer.Listen(ctx, func(topic string, value []byte) error {
		log.Printf("📩 Received event → topic=%s value=%s\n", topic, string(value))
		FraudHandler.Handler(ctx, value)
		return nil
	})

	<-ctx.Done()

	for _, reader := range multiConsumer.Readers {
		if err := reader.Close(); err != nil {
			log.Println("Error closing consumer:", err)
		}
	}

	log.Println("Fraud service stopped")
}

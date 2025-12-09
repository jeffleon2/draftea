package app

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jeffleon2/draftea-payment-service/config"
	handlers "github.com/jeffleon2/draftea-payment-service/internal/handlers"
	"github.com/jeffleon2/draftea-payment-service/internal/models"
	"github.com/jeffleon2/draftea-payment-service/internal/publisher"
	"github.com/jeffleon2/draftea-payment-service/internal/repository/posgrest"
	"github.com/jeffleon2/draftea-payment-service/internal/service"
	"github.com/jeffleon2/draftea-payment-service/internal/subscriber"
	"github.com/sirupsen/logrus"
)

type App struct {
	config *config.Config
	Router *gin.Engine
}

func (a *App) Initialize(cfg *config.Config) {
	a.config = cfg
	db, err := cfg.DB.GormConnect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Payment{}); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	paymentRepo := posgrest.New[models.Payment](db)
	publishTopics := strings.Split(cfg.Kafka.PublishTopics, ",")
	publisher := publisher.NewKafkaPublisher(cfg.Kafka.Brokers, publishTopics, a.config.GetRetryConfig())
	paymentService := service.NewPaymentService(paymentRepo, publisher)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	a.Router = gin.Default()
	a.Router.Use(gin.Recovery())
	a.RegisterRoutes(paymentHandler)

	a.initSubscribers(paymentHandler, publisher, a.config.GetRetryConfig())
}

func (a *App) Run() {
	err := a.Router.Run(fmt.Sprintf(":%s", a.config.APP.PORT))
	if err != nil {
		panic(err)
	}
}

func (a *App) initSubscribers(paymentHandler *handlers.PaymentHandler, publisher *publisher.KafkaPublisher, config config.RetryConfig) {
	brokers := strings.Split(a.config.Kafka.Brokers, ",")
	topics := strings.Split(a.config.Kafka.SubscriberTopics, ",")
	groupID := a.config.Kafka.PaymentConsumerGroup

	consumer := subscriber.NewMultiTopicConsumer(brokers, topics, groupID, publisher, config)

	ctx, _ := context.WithCancel(context.Background())
	go consumer.Listen(ctx, func(topic string, value []byte) error {
		log.Printf("📩 Received message → topic=%s value=%s\n", topic, string(value))
		ctx := context.Background()
		err := paymentHandler.HandleEvents(ctx, topic, value)
		if err != nil {
			logrus.Error(err.Error())
		}
		return nil
	})

}

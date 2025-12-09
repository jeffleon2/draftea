package app

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jeffleon2/draftea-metric-service/config"
	"github.com/jeffleon2/draftea-metric-service/internal/handler"
	"github.com/jeffleon2/draftea-metric-service/internal/metrics"
	"github.com/jeffleon2/draftea-metric-service/internal/subscriber"
	"github.com/sirupsen/logrus"
)

type App struct {
	config *config.Config
	Router *gin.Engine
}

func (a *App) Initialize(cfg *config.Config) {
	a.config = cfg

	metrics.RegisterMetrics()
	metricHandler := handler.NewMetricHandler()
	a.Router = gin.Default()
	a.Router.Use(gin.Recovery())
	a.RegisterRoutes()
	a.initSubscribers(metricHandler)
}

func (a *App) initSubscribers(metricsHandler *handler.MetricsHandler) {
	brokers := strings.Split(a.config.Kafka.Brokers, ",")
	topics := strings.Split(a.config.Kafka.SubscriberTopics, ",")
	groupID := a.config.Kafka.ConsumerGroup

	consumer := subscriber.NewMultiTopicConsumer(brokers, topics, groupID)

	ctx := context.Background()
	go consumer.Listen(ctx, func(topic string, value []byte) {
		log.Printf("📩 Received message → topic=%s value=%s\n", topic, string(value))
		err := metricsHandler.HandleEvents(ctx, topic, value)
		if err != nil {
			logrus.Error(err.Error())
		}
	})

}

func (a *App) Run() {
	err := a.Router.Run(fmt.Sprintf(":%s", a.config.APP.PORT))
	if err != nil {
		panic(err)
	}
}

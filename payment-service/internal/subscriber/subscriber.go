package subscriber

import (
	"context"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/jeffleon2/draftea-payment-service/config"
	"github.com/jeffleon2/draftea-payment-service/internal/models"
	"github.com/jeffleon2/draftea-payment-service/internal/publisher"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	Readers      []*kafka.Reader
	DQLPublisher *publisher.KafkaPublisher
	RetryConfig  config.RetryConfig
}

func NewMultiTopicConsumer(
	brokers []string,
	topics []string,
	groupID string,
	publisher *publisher.KafkaPublisher,
	retryConfig config.RetryConfig,
) *KafkaConsumer {
	readers := make([]*kafka.Reader, len(topics))
	for i, topic := range topics {
		readers[i] = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		})
	}

	return &KafkaConsumer{
		Readers:      readers,
		DQLPublisher: publisher,
		RetryConfig:  retryConfig,
	}
}

func (c *KafkaConsumer) Listen(ctx context.Context, handler func(topic string, value []byte) error) {
	for _, reader := range c.Readers {
		go func(r *kafka.Reader) {
			for {
				msg, err := r.ReadMessage(ctx)
				if err != nil {
					log.Println("Kafka error:", err)
					continue
				}
				c.processMessage(ctx, msg, handler)
			}
		}(reader)
	}
}

func (c *KafkaConsumer) processMessage(ctx context.Context, msg kafka.Message, handler func(topic string, value []byte) error) {
	for attempt := 0; attempt < c.RetryConfig.MaxAttempts; attempt++ {
		err := handler(msg.Topic, msg.Value)
		if err == nil {
			return
		}

		backoff := c.calculateBackoff(attempt)
		log.Printf("Handler error, attempt %d/%d: %v. Retrying in %v", attempt+1, c.RetryConfig.MaxAttempts, err, backoff)
		time.Sleep(backoff)
	}

	log.Printf("Message failed after %d retries: topic=%s, key=%s", c.RetryConfig.MaxAttempts, msg.Topic, string(msg.Key))
	if c.DQLPublisher != nil {
		dlqMessage := models.DLQMessage{
			OriginalTopic: msg.Topic,
			Key:           string(msg.Key),
			Value:         string(msg.Value),
			Timestamp:     time.Now().UTC(),
			Attempts:      c.RetryConfig.MaxAttempts,
		}
		err := c.DQLPublisher.Publish(ctx, models.PaymentsDLQTopic, dlqMessage)
		if err != nil {
			log.Printf("Failed to send message to DLQ: %v", err)
		} else {
			log.Printf("Message sent to DLQ: original topic=%s, key=%s", msg.Topic, string(msg.Key))
		}
	}
}

func (c *KafkaConsumer) calculateBackoff(attempt int) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt))) * c.RetryConfig.BaseDelay

	if delay > c.RetryConfig.MaxDelay {
		delay = c.RetryConfig.MaxDelay
	}

	if c.RetryConfig.Jitter {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.3)
		delay = delay + jitter - time.Duration(float64(delay)*0.15)
	}

	return delay
}

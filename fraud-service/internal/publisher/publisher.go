package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/jeffleon2/draftea-fraud-service/config"
	kafka "github.com/segmentio/kafka-go"
)

// KafkaPublisher manages Kafka message publishing with retry capabilities.
// It maintains a pool of Kafka writers, one per topic, and implements
// exponential backoff retry logic for failed publish attempts.
type KafkaPublisher struct {
	Writers     map[string]*kafka.Writer
	RetryConfig config.RetryConfig
}

// NewKafkaPublisher creates a new KafkaPublisher with writers for the specified topics.
// It initializes default retry configuration values if not provided:
//   - MaxAttempts: 5
//   - BaseDelay: 100ms
//   - MaxDelay: 10s
//
// Each topic gets its own dedicated Kafka writer using LeastBytes balancing strategy.
func NewKafkaPublisher(kafkaURL string, topics []string, retryConfig config.RetryConfig) *KafkaPublisher {
	writers := make(map[string]*kafka.Writer)
	if retryConfig.MaxAttempts == 0 {
		retryConfig.MaxAttempts = 5
	}
	if retryConfig.BaseDelay == 0 {
		retryConfig.BaseDelay = 100 * time.Millisecond
	}
	if retryConfig.MaxDelay == 0 {
		retryConfig.MaxDelay = 10 * time.Second
	}

	for _, t := range topics {
		writers[t] = &kafka.Writer{
			Addr:     kafka.TCP(kafkaURL),
			Topic:    t,
			Balancer: &kafka.LeastBytes{},
		}
	}

	return &KafkaPublisher{
		Writers:     writers,
		RetryConfig: retryConfig,
	}
}

// Publish sends a message to the specified Kafka topic with automatic retry on failure.
// The message is marshaled to JSON before publishing. Returns an error if:
//   - No writer is configured for the topic
//   - JSON marshaling fails
//   - Publishing fails after all retry attempts
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, message interface{}) error {
	writer, ok := p.Writers[topic]
	if !ok {
		return fmt.Errorf("error no writer configured for topic %s", topic)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	msg := kafka.Message{
		Value: data,
	}

	return p.publishWithRetry(ctx, writer, msg, topic)
}

// publishWithRetry attempts to publish a message with exponential backoff retry logic.
// It will retry up to MaxAttempts times, with increasing delays between attempts.
// Returns nil on successful publish, or an error if all attempts fail or context is cancelled.
func (p *KafkaPublisher) publishWithRetry(ctx context.Context, writer *kafka.Writer, msg kafka.Message, topic string) error {
	var lastErr error

	for attempt := 0; attempt < p.RetryConfig.MaxAttempts; attempt++ {
		err := writer.WriteMessages(ctx, msg)
		if err == nil {
			if attempt > 0 {
				fmt.Printf("[Kafka Publisher] Message successfully published to topic '%s' after %d attempts\n", topic, attempt+1)
			}
			return nil
		}

		lastErr = err

		if attempt == p.RetryConfig.MaxAttempts-1 {
			break
		}

		delay := p.calculateBackoff(attempt)

		fmt.Printf("[Kafka Publisher] Retry %d/%d for topic '%s' after %v: %v\n",
			attempt+1, p.RetryConfig.MaxAttempts, topic, delay, err)

		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		}
	}

	return fmt.Errorf("failed to publish message to topic '%s' after %d attempts: %w",
		topic, p.RetryConfig.MaxAttempts, lastErr)
}

// calculateBackoff computes the delay for the next retry attempt using exponential backoff.
// The delay is calculated as: 2^attempt * BaseDelay, capped at MaxDelay.
// If Jitter is enabled, adds random variation (±30%) to prevent thundering herd.
func (p *KafkaPublisher) calculateBackoff(attempt int) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt))) * p.RetryConfig.BaseDelay

	if delay > p.RetryConfig.MaxDelay {
		delay = p.RetryConfig.MaxDelay
	}

	if p.RetryConfig.Jitter {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.3)
		delay = delay + jitter - time.Duration(float64(delay)*0.15)
	}

	return delay
}

package subscriber

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	Readers []*kafka.Reader
}

func NewMultiTopicConsumer(brokers []string, topics []string, groupID string) *KafkaConsumer {
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

	return &KafkaConsumer{Readers: readers}
}

func (c *KafkaConsumer) Listen(ctx context.Context, handler func(topic string, value []byte)) {
	for _, reader := range c.Readers {
		go func(r *kafka.Reader) {
			for {
				msg, err := r.ReadMessage(ctx)
				if err != nil {
					log.Println("Kafka error:", err)
					continue
				}
				handler(msg.Topic, msg.Value)
			}
		}(reader)
	}
}

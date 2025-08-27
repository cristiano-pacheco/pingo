package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Producer interface for producing messages
type Producer interface {
	ProduceMessage(ctx context.Context, topic string, message Message) error
	Close() error
}

// producer implementation
type producer struct {
	writer *kafka.Writer
}

func NewProducer(config Config) Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(config.Address...),
		Balancer: &kafka.LeastBytes{},
		// Producer-specific optimizations
		BatchSize:    100,
		BatchTimeout: 10e6, // 10ms
		RequiredAcks: kafka.RequireOne,
	}

	return &producer{
		writer: writer,
	}
}

func (p *producer) ProduceMessage(ctx context.Context, topic string, message Message) error {
	kafkaHeaders := make([]kafka.Header, len(message.Headers))
	for i, h := range message.Headers {
		kafkaHeaders[i] = kafka.Header{
			Key:   h.Key,
			Value: h.Value,
		}
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     message.Key,
		Value:   message.Value,
		Headers: kafkaHeaders,
	})
}

func (p *producer) Close() error {
	return p.writer.Close()
}

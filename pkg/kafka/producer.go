package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	ProduceMessage(ctx context.Context, message Message) error
	Close() error
}

// producer implementation
type producer struct {
	writer *kafka.Writer
}

func newProducer(config Config, topic string) Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Address...),
		RequiredAcks: kafka.RequireOne,
		Topic:        topic,
	}

	return &producer{
		writer: writer,
	}
}

func (p *producer) ProduceMessage(ctx context.Context, message Message) error {
	kafkaHeaders := make([]kafka.Header, len(message.Headers))
	for i, h := range message.Headers {
		kafkaHeaders[i] = kafka.Header{
			Key:   h.Key,
			Value: h.Value,
		}
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:     message.Key,
		Value:   message.Value,
		Headers: kafkaHeaders,
	})
}

func (p *producer) Close() error {
	return p.writer.Close()
}

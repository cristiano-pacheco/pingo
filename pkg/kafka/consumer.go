package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type MessageHandler func(ctx context.Context, message Message) error

type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
}

type consumer struct {
	config Config
	reader *kafka.Reader
}

func newConsumer(config Config, topic, groupID string) Consumer {
	return &consumer{
		config: config,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: config.Address,
			GroupID: groupID,
			Topic:   topic,
		}),
	}
}

func (c *consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		rawMessage, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		headers := make([]Header, len(rawMessage.Headers))
		for i := range rawMessage.Headers {
			headers[i] = Header{
				Key:   rawMessage.Headers[i].Key,
				Value: rawMessage.Headers[i].Value,
			}
		}

		message := Message{
			Topic:         rawMessage.Topic,
			Partition:     rawMessage.Partition,
			Offset:        rawMessage.Offset,
			HighWaterMark: rawMessage.HighWaterMark,
			Key:           rawMessage.Key,
			Value:         rawMessage.Value,
			Headers:       headers,
			WriterData:    rawMessage.WriterData,
			Time:          rawMessage.Time,
		}

		if err = handler(ctx, message); err != nil {
			// TODO: implement DLQ
			return err
		}
	}
}

func (c *consumer) Close() error {
	return c.reader.Close()
}

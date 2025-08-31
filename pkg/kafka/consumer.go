package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Consume(ctx context.Context, handler func(key, value []byte) error) error
	ReadMessage(ctx context.Context) (Message, error)
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
			Brokers:  config.Address,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
	}
}

func (c *consumer) ReadMessage(ctx context.Context) (Message, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return Message{}, err
	}

	headers := make([]Header, len(m.Headers))
	for i := range m.Headers {
		headers[i] = Header{
			Key:   m.Headers[i].Key,
			Value: m.Headers[i].Value,
		}
	}

	return Message{
		Topic:         m.Topic,
		Partition:     m.Partition,
		Offset:        m.Offset,
		HighWaterMark: m.HighWaterMark,
		Key:           m.Key,
		Value:         m.Value,
		Headers:       headers,
		WriterData:    m.WriterData,
		Time:          m.Time,
	}, nil
}

func (c *consumer) Consume(ctx context.Context, handler func(key, value []byte) error) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		if err := handler(m.Key, m.Value); err != nil {
			return err
		}
	}
}

func (c *consumer) Close() error {
	return c.reader.Close()
}

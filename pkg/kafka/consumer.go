package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
)

// Consumer interface for consuming messages
type Consumer interface {
	ConsumeMessages(ctx context.Context, topic, groupID string, handler func(key, value []byte) error) error
	Close() error
}

// consumer implementation
type consumer struct {
	config  Config
	readers map[string]*kafka.Reader // key: topic+groupID
	mu      sync.RWMutex             // protects readers map
}

func NewConsumer(config Config) Consumer {
	return &consumer{
		config:  config,
		readers: make(map[string]*kafka.Reader),
	}
}

func (c *consumer) ConsumeMessages(ctx context.Context, topic, groupID string, handler func(key, value []byte) error) error {
	reader := c.getOrCreateReader(topic, groupID)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				return err
			}

			if err := handler(msg.Key, msg.Value); err != nil {
				return err
			}
		}
	}
}

func (c *consumer) getOrCreateReader(topic, groupID string) *kafka.Reader {
	key := fmt.Sprintf("%s-%s", topic, groupID)

	c.mu.RLock()
	reader, exists := c.readers[key]
	c.mu.RUnlock()

	if exists {
		return reader
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check pattern
	if reader, exists := c.readers[key]; exists {
		return reader
	}

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.config.Address,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.LastOffset,
	})

	c.readers[key] = reader
	return reader
}

func (c *consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all readers
	for key, reader := range c.readers {
		if err := reader.Close(); err != nil {
			return fmt.Errorf("failed to close reader %s: %w", key, err)
		}
	}

	// Clear readers map
	c.readers = make(map[string]*kafka.Reader)
	return nil
}

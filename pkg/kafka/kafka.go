package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Builder interface {
	BuildProducer(topic string) Producer
	BuildConsumer(topic, groupID string) Consumer
}

type builder struct {
	config Config
}

func NewKafkaBuilder(config Config) Builder {
	return &builder{
		config: config,
	}
}

func MustNewKafkaBuilder(config Config) Builder {
	mustConnection(context.Background(), config)
	return &builder{
		config: config,
	}
}

func (b *builder) BuildProducer(topic string) Producer {
	return newProducer(b.config, topic)
}

func (b *builder) BuildConsumer(topic, groupID string) Consumer {
	return newConsumer(b.config, topic, groupID)
}

func mustConnection(ctx context.Context, config Config) {
	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second,
	}
	conn, err := dialer.DialContext(ctx, "tcp", config.Address[0])
	if err != nil {
		panic(err)
	}
	defer conn.Close()
}

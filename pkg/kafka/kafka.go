package kafka

import (
	"context"
	"log/slog"
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
	logger := slog.Default()
	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second,
	}

	logger.Info("KAFKA: checking connection to Kafka...", slog.Any("address", config.Address))
	conn, err := dialer.DialContext(ctx, "tcp", config.Address[0])
	if err != nil {
		logger.Error("KAFKA: failed to connect to Kafka", slog.Any("error", err))
		panic(err)
	}
	logger.Info("KAFKA: connected to Kafka successfully")
	defer conn.Close()
}

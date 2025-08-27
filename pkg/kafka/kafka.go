package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

type Facade interface {
	Producer(topic string) (*kafka.Writer, error)
	Consumer(topic, groupID string) (*kafka.Reader, error)
}

type facade struct {
	config Config
}

func NewKafkaFacade(config Config) Facade {
	return &facade{
		config: config,
	}
}

func (k *facade) Producer(topic string) (*kafka.Writer, error) {
	w := kafka.Writer{
		Addr:  kafka.TCP(k.config.Address...),
		Topic: topic,
	}

	return &w, nil
}

func (k *facade) Consumer(topic, groupID string) (*kafka.Reader, error) {
	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second,
	}

	readerConfig := kafka.ReaderConfig{
		Brokers: k.config.Address,
		GroupID: groupID,
		Topic:   topic,
		Dialer:  dialer,
	}

	r := kafka.NewReader(readerConfig)

	return r, nil
}

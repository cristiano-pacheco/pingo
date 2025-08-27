package kafka

import (
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
)

func NewKafkaFacade(config config.Config) kafka.Facade {
	kafkaConfig := kafka.Config{
		Address: config.Kafka.Address,
	}
	return kafka.NewKafkaFacade(kafkaConfig)
}

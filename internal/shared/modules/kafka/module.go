package kafka

import "go.uber.org/fx"

var Module = fx.Module("kafka", fx.Provide(NewKafkaFacade))

package kafka

import (
	"context"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

// ConsumerDecorator wraps a kafka consumer with lifecycle management and observability
type ConsumerDecorator struct {
	consumer kafka.Consumer
	logger   logger.Logger
	otel     otel.Otel
	handler  kafka.MessageHandler
	name     string
}

func NewConsumerDecorator(
	consumer kafka.Consumer,
	handler kafka.MessageHandler,
	name string,
	logger logger.Logger,
	otel otel.Otel,
	lc fx.Lifecycle,
) *ConsumerDecorator {
	ctx, cancel := context.WithCancel(context.Background())

	decorator := &ConsumerDecorator{
		consumer: consumer,
		handler:  handler,
		name:     name,
		logger:   logger,
		otel:     otel,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				logger.Info().Msgf("Starting %s...", name)
				if err := decorator.Consume(ctx); err != nil {
					if err == context.Canceled {
						logger.Info().Msgf("%s stopped gracefully", name)
					} else {
						logger.Error().Msgf("%s stopped with error: %v", name, err)
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info().Msgf("Initiating graceful shutdown of %s...", name)

			// Cancel the consumer context to stop consuming new messages
			cancel()

			// Close the kafka consumer
			err := decorator.consumer.Close()
			if err != nil {
				logger.Error().Msgf("failed to close the %s: %v", name, err)
			} else {
				logger.Info().Msgf("%s closed successfully...", name)
			}
			return err
		},
	})

	return decorator
}

// Consume starts consuming messages using the decorated handler
func (d *ConsumerDecorator) Consume(ctx context.Context) error {
	ctx, span := d.otel.StartSpan(ctx, d.name+".Consume")
	defer span.End()

	d.logger.Info().Msgf("%s started consuming messages...", d.name)

	return d.consumer.Consume(ctx, d.handler)
}

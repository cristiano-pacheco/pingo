package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

// MessageProcessor defines the business logic for consuming messages
type MessageProcessor interface {
	Topic() string
	GroupID() string
	ProcessMessage(ctx context.Context, msg kafka.Message) error
}

// ConsumerRunner wires a Consumer with a MessageProcessor
// and manages its lifecycle via Fx.
type ConsumerRunner struct {
	consumer  kafka.Consumer
	processor MessageProcessor
	logger    logger.Logger
}

// NewConsumerRunner creates a ConsumerRunner that automatically
// starts/stops with the Fx lifecycle.
func NewConsumerRunner(
	builder kafka.Builder,
	processor MessageProcessor,
	logger logger.Logger,
	lc fx.Lifecycle,
) *ConsumerRunner {
	runner := &ConsumerRunner{
		processor: processor,
		logger:    logger,
	}

	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			runner.consumer = builder.BuildConsumer(processor.Topic(), processor.GroupID())

			go func() {
				consumerName := fmt.Sprintf("%s:%s", processor.Topic(), processor.GroupID())
				logger.Info().Msgf("Starting consumer %s...", consumerName)

				if err := runner.Run(ctx); err != nil {
					if errors.Is(err, context.Canceled) {
						logger.Info().Msgf("Consumer %s stopped gracefully", consumerName)
					} else {
						logger.Error().Msgf("Consumer %s stopped with error: %v", consumerName, err)
					}
				}
			}()

			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return runner.Stop()
		},
	})

	return runner
}

// Run starts the consumer loop and delegates to the processor.
func (r *ConsumerRunner) Run(ctx context.Context) error {
	return r.consumer.Consume(ctx, func(ctx context.Context, msg kafka.Message) error {
		ctx, span := trace.StartSpan(ctx, "kafka.consumer")
		defer span.End()
		return r.processor.ProcessMessage(ctx, msg)
	})
}

// Stop closes the underlying consumer.
func (r *ConsumerRunner) Stop() error {
	if r.consumer != nil {
		return r.consumer.Close()
	}
	return nil
}

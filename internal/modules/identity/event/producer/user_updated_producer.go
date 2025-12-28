package producer

import (
	"context"
	"encoding/json"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

type UserUpdatedProducerI interface {
	Produce(ctx context.Context, message event.UserUpdatedMessage) error
}

type UserUpdatedProducer struct {
	producer kafka.Producer
}

var _ UserUpdatedProducerI = (*UserUpdatedProducer)(nil)

func NewUserUpdatedProducer(lc fx.Lifecycle,
	logger logger.Logger,
	kafkaBuilder kafka.Builder,
) *UserUpdatedProducer {
	p := UserUpdatedProducer{
		producer: kafkaBuilder.BuildProducer(event.IdentityUserUpdatedTopic),
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			err := p.producer.Close()
			if err != nil {
				logger.Error().Msgf("failed to close producer: %v", err)
			}
			logger.Info().Msg("UserUpdatedProducer closed successfully...")
			return err
		},
	})

	return &p
}

func (p *UserUpdatedProducer) Produce(ctx context.Context, message event.UserUpdatedMessage) error {
	ctx, span := trace.Span(ctx, "UserCreatedProducer.Produce")
	defer span.End()

	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	m := kafka.Message{Value: msg}
	err = p.producer.Produce(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

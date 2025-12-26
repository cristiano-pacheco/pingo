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

type UserCreatedProducer interface {
	Produce(ctx context.Context, message event.UserCreatedMessage) error
}

type userCreatedProducer struct {
	producer kafka.Producer
}

func NewUserCreatedProducer(
	lc fx.Lifecycle,
	logger logger.Logger,
	kafkaBuilder kafka.Builder,
) UserCreatedProducer {
	p := userCreatedProducer{
		producer: kafkaBuilder.BuildProducer(event.IdentityUserCreatedTopic),
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			err := p.producer.Close()
			if err != nil {
				logger.Error().Msgf("failed to close producer: %v", err)
			}
			logger.Info().Msg("UserCreatedProducer closed successfully...")
			return err
		},
	})

	return &p
}

func (p *userCreatedProducer) Produce(ctx context.Context, message event.UserCreatedMessage) error {
	ctx, span := trace.StartSpan(ctx, "UserCreatedProducer.Produce")
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

package producer

import (
	"context"
	"encoding/json"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

type UserAuthenticatedProducer interface {
	Produce(ctx context.Context, message event.UserAuthenticatedMessage) error
}

type userAuthenticatedProducer struct {
	producer kafka.Producer
	otel     otel.Otel
}

func NewUserAuthenticatedProducer(
	lc fx.Lifecycle,
	kafkaFacade kafka.Builder,
	otel otel.Otel,
	logger logger.Logger,
) UserAuthenticatedProducer {
	p := userAuthenticatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserAuthenticatedTopic),
		otel:     otel,
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			err := p.producer.Close()
			if err != nil {
				logger.Error().Msgf("failed to close producer: %v", err)
			}
			logger.Info().Msg("UserAuthenticatedProducer closed successfully...")
			return err
		},
	})

	return &p
}

func (p *userAuthenticatedProducer) Produce(ctx context.Context, message event.UserAuthenticatedMessage) error {
	ctx, span := p.otel.StartSpan(ctx, "UserAuthenticatedProducer.Produce")
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

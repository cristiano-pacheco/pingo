package producer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

type UserCreatedProducer interface {
	Produce(ctx context.Context, userID string) error
}

type userCreatedProducer struct {
	producer kafka.Producer
}

func NewUserCreatedProducer(lc fx.Lifecycle, kafkaFacade kafka.Builder) UserCreatedProducer {
	logger := slog.Default()

	p := userCreatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserCreatedTopic),
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			err := p.producer.Close()
			if err != nil {
				logger.Error("failed to close producer", slog.Any("error", err))
			}
			logger.Info("UserUpdatedProducer closed successfully...")
			return err
		},
	})

	return &p
}

func (p *userCreatedProducer) Produce(ctx context.Context, userID string) error {
	userCreated := event.UserCreatedMessage{
		UserID: userID,
	}
	message, err := json.Marshal(userCreated)
	if err != nil {
		return err
	}
	m := kafka.Message{Value: message}
	err = p.producer.Produce(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

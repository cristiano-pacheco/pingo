package producer

import (
	"context"
	"encoding/json"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

type UserUpdatedProducer interface {
	Produce(ctx context.Context, userID string) error
}

type userUpdatedProducer struct {
	producer kafka.Producer
}

func NewUserUpdatedProducer(lc fx.Lifecycle, kafkaFacade kafka.Builder) UserUpdatedProducer {
	p := userUpdatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserUpdatedTopic),
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return p.Close()
		},
	})

	return &p
}

func (p *userUpdatedProducer) Produce(ctx context.Context, userID string) error {
	userUpdated := event.UserUpdatedMessage{
		UserID: userID,
	}
	message, err := json.Marshal(userUpdated)
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

func (p *userUpdatedProducer) Close() error {
	return p.producer.Close()
}

package producer

import (
	"context"
	"encoding/json"

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
	p := userCreatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserCreatedTopic),
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return p.Close()
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

func (p *userCreatedProducer) Close() error {
	return p.producer.Close()
}

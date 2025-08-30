package producer

import (
	"context"
	"encoding/json"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

type UserAuthenticatedProducer interface {
	Produce(ctx context.Context, userID string) error
}

type userAuthenticatedProducer struct {
	producer kafka.Producer
}

func NewUserAuthenticatedProducer(lc fx.Lifecycle, kafkaFacade kafka.Builder) UserAuthenticatedProducer {
	p := userAuthenticatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserAuthenticatedTopic),
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return p.Close()
		},
	})

	return &p
}

func (p *userAuthenticatedProducer) Produce(ctx context.Context, userID string) error {
	userAuthenticated := event.UserAuthenticatedMessage{
		UserID: userID,
	}
	message, err := json.Marshal(userAuthenticated)
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

func (p *userAuthenticatedProducer) Close() error {
	return p.producer.Close()
}

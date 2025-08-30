package producer

import (
	"context"
	"encoding/json"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
)

type UserAuthenticatedProducer interface {
	Produce(ctx context.Context, userID string) error
	Close()
}

type userAuthenticatedProducer struct {
	producer kafka.Producer
}

func NewUserAuthenticatedProducer(kafkaFacade kafka.Builder) UserAuthenticatedProducer {
	return &userAuthenticatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserAuthenticatedTopic),
	}
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

func (p *userAuthenticatedProducer) Close() {
	p.producer.Close()
}

package producer

import (
	"context"
	"encoding/json"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
)

type UserCreatedProducer interface {
	Produce(ctx context.Context, userID string) error
	Close()
}

type userCreatedProducer struct {
	producer kafka.Producer
}

func NewUserCreatedProducer(kafkaFacade kafka.Builder) UserCreatedProducer {
	return &userCreatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserCreatedTopic),
	}
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

func (p *userCreatedProducer) Close() {
	p.producer.Close()
}

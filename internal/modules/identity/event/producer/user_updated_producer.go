package producer

import (
	"context"
	"encoding/json"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
)

type UserUpdatedProducer interface {
	Produce(ctx context.Context, userID string) error
	Close()
}

type userUpdatedProducer struct {
	producer kafka.Producer
}

func NewUserUpdatedProducer(kafkaFacade kafka.Builder) UserUpdatedProducer {
	return &userUpdatedProducer{
		producer: kafkaFacade.BuildProducer(event.IdentityUserUpdatedTopic),
	}
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

func (p *userUpdatedProducer) Close() {
	p.producer.Close()
}

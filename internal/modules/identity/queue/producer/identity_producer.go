package producer

import (
	"context"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/queue"
	kafka_facade "github.com/cristiano-pacheco/pingo/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

type IdentityProducer interface {
	ProduceUserCreatedEvent(ctx context.Context, userID string) error
	ProduceUserUpdatedEvent(ctx context.Context, userID string) error
	ProduceUserAuthenticatedEvent(ctx context.Context, userID string) error
}

type identityProducer struct {
	kafkaFacade kafka_facade.Facade
}

func NewIdentityProducer(kafkaFacade kafka_facade.Facade) IdentityProducer {
	return &identityProducer{
		kafkaFacade: kafkaFacade,
	}
}

func (p *identityProducer) ProduceUserCreatedEvent(ctx context.Context, userID string) error {
	producer, err := p.kafkaFacade.Producer(queue.IdentityUserCreated)
	if err != nil {
		return err
	}
	defer producer.Close()
	message := kafka.Message{}
	producer.WriteMessages(ctx, message)
	return nil
}

func (p *identityProducer) ProduceUserUpdatedEvent(ctx context.Context, userID string) error {
	producer, err := p.kafkaFacade.Producer(queue.IdentityUserUpdated)
	if err != nil {
		return err
	}
	defer producer.Close()
	return nil
}

func (p *identityProducer) ProduceUserAuthenticatedEvent(ctx context.Context, userID string) error {
	producer, err := p.kafkaFacade.Producer(queue.IdentityUserAuthenticated)
	if err != nil {
		return err
	}
	defer producer.Close()
	return nil
}

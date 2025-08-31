package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

const accountConfirmationTokenExpiration = 24 * time.Hour

type UserCreatedConsumer struct {
	sendEmailConfirmationService service.SendEmailConfirmationService
	oneTimeTokenRepository       repository.OneTimeTokenRepository
	userRepository               repository.UserRepository
	hashService                  service.HashService
	consumer                     kafka.Consumer
	logger                       logger.Logger
	otel                         otel.Otel
}

func NewUserCreatedConsumer(
	sendEmailConfirmationService service.SendEmailConfirmationService,
	oneTimeTokenRepository repository.OneTimeTokenRepository,
	userRepository repository.UserRepository,
	hashService service.HashService,
	consumer kafka.Consumer,
	kafkaBuilder kafka.Builder,
	logger logger.Logger,
	lc fx.Lifecycle,
	otel otel.Otel,
) *UserCreatedConsumer {
	c := UserCreatedConsumer{
		sendEmailConfirmationService: sendEmailConfirmationService,
		oneTimeTokenRepository:       oneTimeTokenRepository,
		userRepository:               userRepository,
		hashService:                  hashService,
		consumer:                     kafkaBuilder.BuildConsumer(event.IdentityUserCreatedTopic, "default"),
		otel:                         otel,
		logger:                       logger,
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			err := c.consumer.Close()
			if err != nil {
				logger.Error().Msgf("failed to close the consumer: %v", err)
			}
			logger.Info().Msg("UserCreatedConsumer closed successfully...")
			return err
		},
	})

	return &c
}

func (c *UserCreatedConsumer) Consume() error {
	ctx, span := c.otel.StartSpan(context.Background(), "UserCreatedConsumer.Consume")
	defer span.End()

	rawMessage, err := c.consumer.ReadMessage(ctx)
	if err != nil {
		return err
	}

	var message event.UserCreatedMessage
	if err := json.Unmarshal(rawMessage.Value, &message); err != nil {
		return err
	}

	if message.UserID == 0 {
		c.logger.Error().Msg("invalid user ID")
		return errors.New("invalid user ID")
	}

	user, err := c.userRepository.FindByID(ctx, message.UserID)
	if err != nil {
		return err
	}

	token, err := c.hashService.GenerateRandomBytes()
	if err != nil {
		c.logger.Error().Msgf("error generating random bytes: %v", err)
		return err
	}

	oneTimeToken := model.OneTimeTokenModel{
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(accountConfirmationTokenExpiration),
		TokenType: enum.TokenTypeAccountConfirmation,
		TokenHash: token,
	}

	_, err = c.oneTimeTokenRepository.Create(ctx, oneTimeToken)
	if err != nil {
		c.logger.Error().Msgf("error creating one-time token: %v", err)
		return err
	}

	sendEmailConfirmationInput := service.SendEmailConfirmationInput{
		UserModel:             user,
		ConfirmationTokenHash: token,
	}
	err = c.sendEmailConfirmationService.Execute(ctx, sendEmailConfirmationInput)
	if err != nil {
		c.logger.Error().Msgf("error sending account confirmation email: %v", err)
		return err
	}

	return nil
}

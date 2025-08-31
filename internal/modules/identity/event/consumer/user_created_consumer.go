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
	kafkaBuilder kafka.Builder,
	logger logger.Logger,
	lc fx.Lifecycle,
	otel otel.Otel,
) *UserCreatedConsumer {
	ctx, cancel := context.WithCancel(context.Background())

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
		OnStart: func(_ context.Context) error {
			go func() {
				logger.Info().Msg("Starting UserCreatedConsumer...")
				if err := c.Consume(ctx); err != nil {
					if err == context.Canceled {
						logger.Info().Msg("UserCreatedConsumer stopped gracefully")
					} else {
						logger.Error().Msgf("UserCreatedConsumer stopped with error: %v", err)
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info().Msg("Initiating graceful shutdown of UserCreatedConsumer...")

			// Cancel the consumer context to stop consuming new messages
			cancel()

			// Close the kafka consumer
			err := c.consumer.Close()
			if err != nil {
				logger.Error().Msgf("failed to close the consumer: %v", err)
			} else {
				logger.Info().Msg("UserCreatedConsumer closed successfully...")
			}
			return err
		},
	})

	return &c
}

func (c *UserCreatedConsumer) Consume(ctx context.Context) error {
	ctx, span := c.otel.StartSpan(ctx, "UserCreatedConsumer.Consume")
	defer span.End()

	c.logger.Info().Msg("UserCreatedConsumer started consuming messages...")

	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *UserCreatedConsumer) handleMessage(ctx context.Context, message kafka.Message) error {
	ctx, span := c.otel.StartSpan(ctx, "UserCreatedConsumer.handleMessage")
	defer span.End()

	var userCreatedMessage event.UserCreatedMessage
	if err := json.Unmarshal(message.Value, &userCreatedMessage); err != nil {
		c.logger.Error().Msgf("error unmarshaling message: %v", err)
		return err
	}

	if userCreatedMessage.UserID == 0 {
		c.logger.Error().Msg("invalid user ID")
		return errors.New("invalid user ID")
	}

	user, err := c.userRepository.FindByID(ctx, userCreatedMessage.UserID)
	if err != nil {
		c.logger.Error().Msgf("error finding user by ID: %v", err)
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

	c.logger.Info().Msgf("Successfully processed user created event for user ID: %d", userCreatedMessage.UserID)
	return nil
}

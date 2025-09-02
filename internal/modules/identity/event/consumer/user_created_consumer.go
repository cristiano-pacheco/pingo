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
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/kafka"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	pkgkafka "github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

const accountConfirmationTokenExpiration = 24 * time.Hour

// UserCreatedConsumer handles user creation events
type UserCreatedConsumer struct {
	sendEmailConfirmationService service.SendEmailConfirmationService
	oneTimeTokenRepository       repository.OneTimeTokenRepository
	userRepository               repository.UserRepository
	hashService                  service.HashService
	logger                       logger.Logger
	otel                         otel.Otel
}

// NewUserCreatedConsumer creates a new user created event consumer using the decorator pattern
func NewUserCreatedConsumer(
	sendEmailConfirmationService service.SendEmailConfirmationService,
	oneTimeTokenRepository repository.OneTimeTokenRepository,
	userRepository repository.UserRepository,
	hashService service.HashService,
	kafkaBuilder pkgkafka.Builder,
	logger logger.Logger,
	lc fx.Lifecycle,
	otel otel.Otel,
) *kafka.ConsumerDecorator {
	userConsumer := &UserCreatedConsumer{
		sendEmailConfirmationService: sendEmailConfirmationService,
		oneTimeTokenRepository:       oneTimeTokenRepository,
		userRepository:               userRepository,
		hashService:                  hashService,
		logger:                       logger,
		otel:                         otel,
	}

	consumer := kafkaBuilder.BuildConsumer(event.IdentityUserCreatedTopic, "default")

	return kafka.NewConsumerDecorator(
		consumer,
		userConsumer.processMessage,
		"identity.user.created",
		logger,
		otel,
		lc,
	)
}

func (c *UserCreatedConsumer) processMessage(ctx context.Context, message pkgkafka.Message) error {
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

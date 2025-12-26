package consumer

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
)

const (
	verificationCodeTTL = 10 * time.Minute
	maxRandomNumber     = 1000000
)

type UserAuthenticatedConsumer struct {
	sendEmailVerificationCodeService service.SendEmailVerificationCodeService
	oneTimeTokenRepository           repository.OneTimeTokenRepository
	userRepository                   repository.UserRepository
	hashService                      service.HashService
	logger                           logger.Logger
}

func NewUserAuthenticatedConsumer(
	sendEmailVerificationCodeService service.SendEmailVerificationCodeService,
	oneTimeTokenRepository repository.OneTimeTokenRepository,
	userRepository repository.UserRepository,
	hashService service.HashService,
	logger logger.Logger,
) *UserAuthenticatedConsumer {
	return &UserAuthenticatedConsumer{
		oneTimeTokenRepository:           oneTimeTokenRepository,
		userRepository:                   userRepository,
		hashService:                      hashService,
		sendEmailVerificationCodeService: sendEmailVerificationCodeService,
		logger:                           logger,
	}
}

func (c *UserAuthenticatedConsumer) Topic() string {
	return event.IdentityUserAuthenticatedTopic
}

func (c *UserAuthenticatedConsumer) GroupID() string {
	return "default"
}

func (c *UserAuthenticatedConsumer) ProcessMessage(ctx context.Context, message kafka.Message) error {
	ctx, span := trace.StartSpan(ctx, "AuthLoginUseCase.Execute")
	defer span.End()

	var userAuthenticatedMessage event.UserAuthenticatedMessage
	if err := json.Unmarshal(message.Value, &userAuthenticatedMessage); err != nil {
		c.logger.Error().Msgf("error unmarshaling message: %v", err)
		return err
	}

	user, err := c.userRepository.FindByID(ctx, userAuthenticatedMessage.UserID)
	if err != nil {
		c.logger.Error().Msgf("error finding user by ID %d: %v", userAuthenticatedMessage.UserID, err)
		return err
	}

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	if err = c.oneTimeTokenRepository.Delete(ctx, user.ID, emailVerificationType); err != nil &&
		!errors.Is(err, errs.ErrRecordNotFound) {
		c.logger.Error().Msgf("error deleting verification codes for user ID %d: %v", user.ID, err)
		return err
	}

	n, err := rand.Int(rand.Reader, big.NewInt(maxRandomNumber))
	if err != nil {
		c.logger.Error().Msgf("error generating verification code: %v", err)
		return err
	}

	code := fmt.Sprintf("%06d", n.Int64())
	tokenHash, err := c.hashService.GenerateFromPassword([]byte(code))
	if err != nil {
		c.logger.Error().Msgf("error hashing verification code for user ID %d: %v", user.ID, err)
		return err
	}

	tokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	oneTimeToken := model.OneTimeTokenModel{
		UserID:    user.ID,
		TokenHash: tokenHash,
		TokenType: tokenType.String(),
		ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
		CreatedAt: time.Now().UTC(),
	}

	if _, err = c.oneTimeTokenRepository.Create(ctx, oneTimeToken); err != nil {
		c.logger.Error().Msgf("error creating one-time token for user ID %d: %v", user.ID, err)
		return err
	}

	sendEmailVerificationCodeInput := service.SendEmailVerificationCodeInput{UserID: user.ID, Code: code}
	if err = c.sendEmailVerificationCodeService.Execute(ctx, sendEmailVerificationCodeInput); err != nil {
		c.logger.Error().Msgf("error sending verification code email for user ID %d: %v", user.ID, err)
		return err
	}

	c.logger.Info().
		Msgf("Successfully processed user authenticated event for user ID: %d", userAuthenticatedMessage.UserID)
	return nil
}

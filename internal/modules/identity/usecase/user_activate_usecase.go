package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/cache"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type UserActivateInput struct {
	UserID uint64 `validate:"required"`
	Token  string `validate:"required"`
}

type UserActivateUseCase struct {
	oneTimeTokenRepository repository.OneTimeTokenRepository
	userRepository         repository.UserRepository
	userActivatedCache     cache.UserActivatedCache
	validate               validator.Validate
	logger                 logger.Logger
	otel                   otel.Otel
}

func NewUserActivateUseCase(
	oneTimeTokenRepository repository.OneTimeTokenRepository,
	userRepository repository.UserRepository,
	userActivatedCache cache.UserActivatedCache,
	validate validator.Validate,
	logger logger.Logger,
	otel otel.Otel,
) *UserActivateUseCase {
	return &UserActivateUseCase{
		oneTimeTokenRepository: oneTimeTokenRepository,
		userRepository:         userRepository,
		userActivatedCache:     userActivatedCache,
		validate:               validate,
		logger:                 logger,
		otel:                   otel,
	}
}

func (uc *UserActivateUseCase) Execute(ctx context.Context, input UserActivateInput) error {
	ctx, span := uc.otel.StartSpan(ctx, "UserActivateUseCase.Execute")
	defer span.End()

	err := uc.validate.Struct(input)
	if err != nil {
		return err
	}

	token, err := base64.StdEncoding.DecodeString(input.Token)
	if err != nil {
		uc.logger.Error().Msgf("Failed to decode token: %v", err)
		return err
	}

	user, err := uc.userRepository.FindByID(ctx, input.UserID)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		uc.logger.Error().
			Msgf("Failed to find user with confirmation token for the user_id: %d, error: %v", user.ID, err)
		return err
	}

	if user.Status != enum.UserStatusPending {
		return errs.ErrUserNotInPendingStatus
	}

	confirmationTokenType, err := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	if err != nil {
		return err
	}

	oneTimeToken, err := uc.oneTimeTokenRepository.Find(ctx, user.ID, confirmationTokenType)
	if err != nil {
		uc.logger.Error().Msgf("Failed to find one-time token for the user_id: %d, error: %v", user.ID, err)
		return err
	}

	if oneTimeToken.ID == 0 || string(oneTimeToken.TokenHash) != string(token) {
		return errs.ErrInvalidAccountConfirmationToken
	}

	now := time.Now().UTC()
	user.ConfirmedAt = &now
	user.UpdatedAt = now
	user.Status = enum.UserStatusActive
	err = uc.userRepository.Update(ctx, user)
	if err != nil {
		uc.logger.Error().Msgf("Failed to update user confirmation status for the user_id: %d, error %v", user.ID, err)
		return err
	}

	err = uc.oneTimeTokenRepository.Delete(ctx, user.ID, confirmationTokenType)
	if err != nil {
		uc.logger.Error().Msgf("Failed to delete one-time token for the user_id: %d, error: %v", user.ID, err)
		return err
	}

	// Set user as activated in cache for fast lookup
	err = uc.userActivatedCache.Set(user.ID)
	if err != nil {
		// Log the error but don't fail the request since the user is already activated in the database
		uc.logger.Warn().Msgf("Failed to set user in activation cache for user_id: %d, error: %v", user.ID, err)
	}

	return nil
}

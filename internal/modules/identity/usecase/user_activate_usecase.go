package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type UserActivateInput struct {
	Token string `validate:"required"`
}

type UserActivateUseCase struct {
	userRepository repository.UserRepository
	validate       validator.Validate
	logger         logger.Logger
}

func NewUserActivateUseCase(
	userRepository repository.UserRepository,
	validate validator.Validate,
	logger logger.Logger,
) *UserActivateUseCase {
	return &UserActivateUseCase{userRepository, validate, logger}
}

func (uc *UserActivateUseCase) Execute(ctx context.Context, input UserActivateInput) error {
	ctx, span := otel.Trace().StartSpan(ctx, "UserActivateUseCase.Execute")
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

	user, err := uc.userRepository.FindPendingConfirmation(ctx, token)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		uc.logger.Error().
			Msgf("Failed to find user with confirmation token for the user_id: %d, error: %v", user.ID, err)
		return err
	}

	if user.ID == 0 {
		return errs.ErrInvalidAccountConfirmationToken
	}

	now := time.Now().UTC()
	user.ConfirmedAt = &now
	user.UpdatedAt = now
	user.ConfirmationToken = []byte{}
	user.Status = enum.UserStatusActive
	err = uc.userRepository.Update(ctx, user)
	if err != nil {
		uc.logger.Error().Msgf("Failed to update user confirmation status for the user_id: %d, error %v", user.ID, err)
		return err
	}

	return nil
}

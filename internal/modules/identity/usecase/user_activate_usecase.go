package usecase

import (
	"context"
	"errors"
	"time"

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

	user, err := uc.userRepository.FindPendingConfirmation(ctx, []byte(input.Token))
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		return err
	}

	if user.ID == 0 {
		return errs.ErrInvalidAccountConfirmationToken
	}

	now := time.Now().UTC()
	user.ConfirmedAt = &now
	user.UpdatedAt = now
	user.ConfirmationToken = nil
	err = uc.userRepository.Update(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

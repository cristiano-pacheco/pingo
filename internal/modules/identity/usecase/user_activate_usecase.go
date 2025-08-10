package usecase

import (
	"context"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type UserActivateUseCase interface {
	Execute(ctx context.Context, input UserActivateInput) error
}

type UserActivateInput struct {
	Token string `validate:"required"`
}

type userActivateUseCase struct {
	userRepository repository.UserRepository
	validate       validator.Validate
	logger         logger.Logger
}

func NewUserActivateUseCase(
	userRepository repository.UserRepository,
	validate validator.Validate,
	logger logger.Logger,
) UserActivateUseCase {
	return &userActivateUseCase{userRepository, validate, logger}
}

func (uc *userActivateUseCase) Execute(ctx context.Context, input UserActivateInput) error {
	ctx, span := otel.Trace().StartSpan(ctx, "UserActivateUseCase.Execute")
	defer span.End()

	err := uc.validate.Struct(input)
	if err != nil {
		return err
	}

	user, err := uc.userRepository.FindByConfirmationToken(ctx, []byte(input.Token))
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	user.ConfirmedAt = &now
	user.UpdatedAt = now
	err = uc.userRepository.Update(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

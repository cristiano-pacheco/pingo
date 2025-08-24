package usecase

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	identity_validator "github.com/cristiano-pacheco/pingo/internal/modules/identity/validator"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type UserUpdateInput struct {
	UserID    uint64 `validate:"required"`
	FirstName string `validate:"required,min=3,max=255"`
	LastName  string `validate:"required,min=3,max=255"`
	Password  string `validate:"omitempty,min=8"`
}

type UserUpdateOutput struct {
	UserID    uint64
	FirstName string
	LastName  string
	Email     string
}

type UserUpdateUseCase struct {
	hashService       service.HashService
	userRepository    repository.UserRepository
	validate          validator.Validate
	passwordValidator identity_validator.PasswordValidator
	logger            logger.Logger
	otel              otel.Otel
}

func NewUserUpdateUseCase(
	hashService service.HashService,
	userRepo repository.UserRepository,
	validate validator.Validate,
	passwordValidator identity_validator.PasswordValidator,
	logger logger.Logger,
	otel otel.Otel,
) *UserUpdateUseCase {
	return &UserUpdateUseCase{
		hashService,
		userRepo,
		validate,
		passwordValidator,
		logger,
		otel,
	}
}

func (uc *UserUpdateUseCase) Execute(ctx context.Context, input UserUpdateInput) error {
	ctx, span := uc.otel.StartSpan(ctx, "UserUpdateUseCase.Execute")
	defer span.End()

	err := uc.validate.Struct(input)
	if err != nil {
		return err
	}

	user, err := uc.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, shared_errs.ErrRecordNotFound) {
			return errs.ErrUserNotFound
		}
		uc.logger.Error().Msgf("error finding user by id: %v", err)
		return err
	}

	if input.Password != "" {
		if err = uc.passwordValidator.Validate(input.Password); err != nil {
			return err
		}
		var passwordHash []byte
		passwordHash, err = uc.hashService.GenerateFromPassword([]byte(input.Password))
		if err != nil {
			uc.logger.Error().Msgf("error generating password hash: %v", err)
			return err
		}
		user.PasswordHash = passwordHash
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName

	err = uc.userRepository.Update(ctx, user)
	if err != nil {
		uc.logger.Error().Msgf("error updating user: %v", err)
		return err
	}

	return nil
}

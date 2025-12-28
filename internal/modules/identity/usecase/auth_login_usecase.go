package usecase

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/producer"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"

	"github.com/cristiano-pacheco/go-otel/trace"
)

type AuthLoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type AuthLoginOutput struct {
	UserID uint64
}

type AuthLoginUseCase struct {
	userAuthenticatedProducer producer.UserAuthenticatedProducer
	userRepository            repository.UserRepository
	hashService               service.HashService
	validator                 validator.Validate
	logger                    logger.Logger
}

func NewAuthLoginUseCase(
	userAuthenticatedProducer producer.UserAuthenticatedProducer,
	userRepository repository.UserRepository,
	validator validator.Validate,
	hashService service.HashService,
	logger logger.Logger,
) *AuthLoginUseCase {
	return &AuthLoginUseCase{
		userAuthenticatedProducer: userAuthenticatedProducer,
		userRepository:            userRepository,
		validator:                 validator,
		hashService:               hashService,
		logger:                    logger,
	}
}

func (u *AuthLoginUseCase) Execute(ctx context.Context, input AuthLoginInput) (AuthLoginOutput, error) {
	ctx, span := trace.Span(ctx, "AuthLoginUseCase.Execute")
	defer span.End()
	if err := u.validator.Struct(input); err != nil {
		return AuthLoginOutput{}, err
	}

	user, err := u.userRepository.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		u.logger.Error().Msgf("error finding by email %v", err)
		return AuthLoginOutput{}, err
	}

	if user.ID == 0 {
		return AuthLoginOutput{}, errs.ErrInvalidCredentials
	}

	if user.Status != enum.UserStatusActive {
		return AuthLoginOutput{}, errs.ErrUserIsNotActive
	}

	if err = u.hashService.CompareHashAndPassword(user.PasswordHash, []byte(input.Password)); err != nil {
		return AuthLoginOutput{}, errs.ErrInvalidCredentials
	}

	message := event.UserAuthenticatedMessage{UserID: user.ID}
	err = u.userAuthenticatedProducer.Produce(ctx, message)
	if err != nil {
		u.logger.Error().Msgf("error producing user authenticated message for user id: %v, error: %v", user.ID, err)
		return AuthLoginOutput{}, err
	}

	return AuthLoginOutput{UserID: user.ID}, nil
}

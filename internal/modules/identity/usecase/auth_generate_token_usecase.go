package usecase

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type AuthGenerateTokenUseCase struct {
	oneTimeTokenRepository repository.OneTimeTokenRepository
	userRepository         repository.UserRepository
	tokenService           service.TokenService
	hashService            service.HashService
	validator              validator.Validate
	logger                 logger.Logger
}

func NewAuthGenerateTokenUseCase(
	oneTimeTokenRepository repository.OneTimeTokenRepository,
	userRepository repository.UserRepository,
	tokenService service.TokenService,
	hashService service.HashService,
	validator validator.Validate,
	logger logger.Logger,
) *AuthGenerateTokenUseCase {
	return &AuthGenerateTokenUseCase{
		oneTimeTokenRepository: oneTimeTokenRepository,
		userRepository:         userRepository,
		tokenService:           tokenService,
		hashService:            hashService,
		validator:              validator,
		logger:                 logger,
	}
}

type GenerateTokenInput struct {
	UserID uint64 `validate:"required"`
	Code   string `validate:"required"`
}

type GenerateTokenOutput struct {
	Token string
}

func (uc *AuthGenerateTokenUseCase) Execute(
	ctx context.Context,
	input GenerateTokenInput,
) (GenerateTokenOutput, error) {
	ctx, span := otel.Trace().StartSpan(ctx, "AuthGenerateTokenUseCase.Execute")
	defer span.End()

	output := GenerateTokenOutput{}

	err := uc.validator.Struct(input)
	if err != nil {
		return output, err
	}

	user, err := uc.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, shared_errs.ErrRecordNotFound) {
			return output, errs.ErrInvalidCredentials
		}
		uc.logger.Error().Msgf("error finding user by ID %d: %v", input.UserID, err)
		return output, err
	}

	if user.Status != enum.UserStatusActive {
		return output, errs.ErrUserIsNotActive
	}

	loginVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	oneTimeToken, err := uc.oneTimeTokenRepository.Find(ctx, input.UserID, loginVerificationType)
	if err != nil {
		if errors.Is(err, shared_errs.ErrRecordNotFound) {
			return output, errs.ErrInvalidCredentials
		}
		uc.logger.Error().Msgf("error finding one-time token for the user %d: %v", input.UserID, err)
		return output, err
	}

	err = uc.hashService.CompareHashAndPassword(oneTimeToken.TokenHash, []byte(input.Code))
	if err != nil {
		return output, err
	}

	err = uc.oneTimeTokenRepository.Delete(ctx, input.UserID, loginVerificationType)
	if err != nil {
		uc.logger.Error().Msgf("error deleting one-time token for the user %d: %v", input.UserID, err)
		return output, err
	}

	token, err := uc.tokenService.GenerateJWT(ctx, user)
	if err != nil {
		return output, err
	}

	return GenerateTokenOutput{Token: token}, nil
}

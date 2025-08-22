package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
	"github.com/samber/lo"
)

type AuthGenerateTokenUseCase struct {
	verificationCodeRepository repository.VerificationCodeRepository
	userRepository             repository.UserRepository
	validator                  validator.Validate
	tokenService               service.TokenService
	logger                     logger.Logger
}

func NewAuthGenerateTokenUseCase(
	verificationCodeRepository repository.VerificationCodeRepository,
	userRepository repository.UserRepository,
	validator validator.Validate,
	tokenService service.TokenService,
	logger logger.Logger,
) *AuthGenerateTokenUseCase {
	return &AuthGenerateTokenUseCase{
		verificationCodeRepository: verificationCodeRepository,
		userRepository:             userRepository,
		validator:                  validator,
		tokenService:               tokenService,
		logger:                     logger,
	}
}

type GenerateTokenInput struct {
	UserID uint64 `validate:"required"`
	Code   string `validate:"required,numeric,len=6"`
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

	verificationCode, err := uc.verificationCodeRepository.FindByUserAndCode(ctx, input.UserID, input.Code)
	if err != nil {
		if errors.Is(err, shared_errs.ErrRecordNotFound) {
			return output, errs.ErrInvalidCredentials
		}
		uc.logger.Error().Msgf("error finding verification for the user %d: %v", input.UserID, err)
		return output, err
	}

	verificationCode.UsedAt = lo.ToPtr(time.Now().UTC())
	err = uc.verificationCodeRepository.Update(ctx, verificationCode)
	if err != nil {
		uc.logger.Error().Msgf("error updating verification code %s: %v", input.Code, err)
		return output, err
	}

	token, err := uc.tokenService.GenerateJWT(ctx, user)
	if err != nil {
		return output, err
	}

	return GenerateTokenOutput{Token: token}, nil
}

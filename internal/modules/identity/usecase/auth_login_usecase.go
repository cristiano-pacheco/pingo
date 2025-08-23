package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

const (
	verificationCodeTTL = 10 * time.Minute
	maxRandomNumber     = 1000000
)

type AuthLoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type AuthLoginOutput struct {
	UserID uint64
}

type AuthLoginUseCase struct {
	oneTimeTokenRepository           repository.OneTimeTokenRepository
	userRepository                   repository.UserRepository
	validator                        validator.Validate
	hashService                      service.HashService
	sendEmailVerificationCodeService service.SendEmailVerificationCodeService
	logger                           logger.Logger
}

func NewAuthLoginUseCase(
	oneTimeTokenRepository repository.OneTimeTokenRepository,
	userRepository repository.UserRepository,
	validator validator.Validate,
	hashService service.HashService,
	sendEmailVerificationCodeService service.SendEmailVerificationCodeService,
	logger logger.Logger,
) *AuthLoginUseCase {
	return &AuthLoginUseCase{
		userRepository:                   userRepository,
		oneTimeTokenRepository:           oneTimeTokenRepository,
		validator:                        validator,
		hashService:                      hashService,
		sendEmailVerificationCodeService: sendEmailVerificationCodeService,
		logger:                           logger,
	}
}

func (u *AuthLoginUseCase) Execute(ctx context.Context, input AuthLoginInput) (AuthLoginOutput, error) {
	ctx, span := otel.Trace().StartSpan(ctx, "AuthLoginUseCase.Execute")
	defer span.End()

	if err := u.validator.Struct(input); err != nil {
		return AuthLoginOutput{}, err
	}

	user, err := u.userRepository.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		u.logger.Error().Msgf("error finding by email %v", err)
		return AuthLoginOutput{}, err
	}

	if err = u.validateUserForLogin(user, input.Password); err != nil {
		return AuthLoginOutput{}, err
	}

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	err = u.oneTimeTokenRepository.Delete(ctx, user.ID, emailVerificationType)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		u.logger.Error().Msgf("error deleting verification codes for user ID %d: %v", user.ID, err)
		return AuthLoginOutput{}, err
	}

	n, err := rand.Int(rand.Reader, big.NewInt(maxRandomNumber))
	if err != nil {
		u.logger.Error().Msgf("error generating verification code: %v", err)
		return AuthLoginOutput{}, err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	tokenHash, err := u.hashService.GenerateFromPassword([]byte(code))
	if err != nil {
		u.logger.Error().Msgf("error hashing verification code for user ID %d: %v", user.ID, err)
		return AuthLoginOutput{}, err
	}
	oneTimeToken := model.OneTimeTokenModel{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
		CreatedAt: time.Now().UTC(),
	}

	_, err = u.oneTimeTokenRepository.Create(ctx, oneTimeToken)
	if err != nil {
		u.logger.Error().Msgf("error creating one-time token for user ID %d: %v", user.ID, err)
		return AuthLoginOutput{}, err
	}

	sendEmailVerificationCodeInput := service.SendEmailVerificationCodeInput{UserID: user.ID, Code: code}
	err = u.sendEmailVerificationCodeService.Execute(ctx, sendEmailVerificationCodeInput)
	if err != nil {
		u.logger.Error().Msgf("error sending verification code email for user ID %d: %v", user.ID, err)
		return AuthLoginOutput{}, err
	}

	return AuthLoginOutput{UserID: user.ID}, nil
}

func (u *AuthLoginUseCase) validateUserForLogin(user model.UserModel, password string) error {
	if user.ID == 0 {
		return errs.ErrInvalidCredentials
	}

	if user.Status != "active" {
		return errs.ErrUserIsNotActive
	}

	if err := u.hashService.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return errs.ErrInvalidCredentials
	}

	return nil
}

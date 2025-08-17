package usecase

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	identity_validator "github.com/cristiano-pacheco/pingo/internal/modules/identity/validator"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type UserCreateInput struct {
	FirstName string `validate:"required,min=3,max=255"`
	LastName  string `validate:"required,min=3,max=255"`
	Password  string `validate:"required,min=8"`
	Email     string `validate:"required,email"`
}

type UserCreateOutput struct {
	FirstName string
	LastName  string
	Email     string
	UserID    uint64
}

type UserCreateUseCase struct {
	sendEmailConfirmationService service.SendEmailConfirmationService
	hashService                  service.HashService
	userRepository               repository.UserRepository
	validate                     validator.Validate
	passwordValidator            identity_validator.PasswordValidator
	logger                       logger.Logger
}

func NewUserCreateUseCase(
	sendEmailConfirmationService service.SendEmailConfirmationService,
	hashService service.HashService,
	userRepo repository.UserRepository,
	validate validator.Validate,
	passwordValidator identity_validator.PasswordValidator,
	logger logger.Logger,
) *UserCreateUseCase {
	return &UserCreateUseCase{
		sendEmailConfirmationService,
		hashService,
		userRepo,
		validate,
		passwordValidator,
		logger,
	}
}

func (uc *UserCreateUseCase) Execute(ctx context.Context, input UserCreateInput) (UserCreateOutput, error) {
	ctx, span := otel.Trace().StartSpan(ctx, "UserCreateUseCase.Execute")
	defer span.End()

	output := UserCreateOutput{}

	err := uc.validate.Struct(input)
	if err != nil {
		return output, err
	}

	if err := uc.passwordValidator.Validate(input.Password); err != nil {
		return output, err
	}

	user, err := uc.userRepository.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, shared_errs.ErrRecordNotFound) {
		uc.logger.Error().Msgf("error finding user by email: %v", err)
		return output, err
	}

	if user.ID != 0 {
		return output, errs.ErrEmailAlreadyInUse
	}

	token, err := uc.hashService.GenerateRandomBytes()
	if err != nil {
		uc.logger.Error().Msgf("error generating random bytes: %v", err)
		return output, err
	}

	passwordHash, err := uc.hashService.GenerateFromPassword([]byte(input.Password))
	if err != nil {
		uc.logger.Error().Msgf("error generating password hash: %v", err)
		return output, err
	}

	pendingUserStatus := enum.UserStatusPending
	userModel := model.UserModel{
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		Email:             input.Email,
		PasswordHash:      passwordHash,
		Status:            pendingUserStatus,
		ConfirmationToken: token,
	}

	createdUser, err := uc.userRepository.Create(ctx, userModel)
	if err != nil {
		uc.logger.Error().Msgf("error creating user: %v", err)
		return output, err
	}

	err = uc.sendEmailConfirmationService.Execute(ctx, createdUser.ID)
	if err != nil {
		uc.logger.Error().Msgf("error sending account confirmation email: %v", err)
		return output, err
	}

	output = UserCreateOutput{
		UserID:    createdUser.ID,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Email:     createdUser.Email,
	}

	return output, nil
}

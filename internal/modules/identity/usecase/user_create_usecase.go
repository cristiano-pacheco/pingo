package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/identity/service"
	"github.com/cristiano-pacheco/pingo/internal/shared/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/validator"
)

type UserCreateUseCase interface {
	Execute(ctx context.Context, input UserCreateInput) (UserCreateOutput, error)
}

type UserCreateInput struct {
	Name     string `validate:"required,min=3,max=255"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

type UserCreateOutput struct {
	Name   string
	Email  string
	UserID uint64
}

type userCreateUseCase struct {
	sendEmailConfirmationService service.SendEmailConfirmationService
	hashService                  service.HashService
	userRepository               repository.UserRepository
	validate                     validator.Validate
	logger                       logger.Logger
}

func NewUserCreateUseCase(
	sendEmailConfirmationService service.SendEmailConfirmationService,
	hashService service.HashService,
	userRepo repository.UserRepository,
	validate validator.Validate,
	logger logger.Logger,
) UserCreateUseCase {
	return &userCreateUseCase{
		sendEmailConfirmationService,
		hashService,
		userRepo,
		validate,
		logger,
	}
}

func (uc *userCreateUseCase) Execute(ctx context.Context, input UserCreateInput) (UserCreateOutput, error) {
	ctx, span := otel.Trace().StartSpan(ctx, "UserCreateUseCase.Execute")
	defer span.End()

	output := UserCreateOutput{}

	err := uc.validate.Struct(input)
	if err != nil {
		return output, err
	}

	user, err := uc.userRepository.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, kernel_errs.ErrNotFound) {
		uc.logger.Error("error finding user by email", "error", err)
		return output, err
	}

	if user.ID() != 0 {
		return output, errs.ErrEmailAlreadyInUse
	}

	ph, err := uc.hashService.GenerateFromPassword([]byte(input.Password))
	if err != nil {
		message := "error generating password hash"
		uc.logger.Error(message, "error", err)
		return output, err
	}

	token, err := uc.hashService.GenerateRandomBytes()
	if err != nil {
		message := "error generating random bytes"
		uc.logger.Error(message, "error", err)
		return output, err
	}

	// encode the token
	confirmationToken := base64.StdEncoding.EncodeToString(token)
	confirmationExpiresAt := time.Now().UTC().Add(time.Hour * 24)

	userModel, err := model.CreateUserModel(
		input.Name,
		input.Email,
		string(ph),
		confirmationToken,
		confirmationExpiresAt,
	)
	if err != nil {
		message := "error creating user model"
		uc.logger.Error(message, "error", err)
		return output, err
	}

	newUserModel, err := uc.userRepository.Create(ctx, userModel)
	if err != nil {
		message := "error creating user"
		uc.logger.Error(message, "error", err)
		return output, err
	}

	err = uc.sendEmailConfirmationService.Execute(ctx, newUserModel.ID())
	if err != nil {
		message := "error sending account confirmation email"
		uc.logger.Error(message, "error", err)
		return output, err
	}

	output = UserCreateOutput{
		UserID: newUserModel.ID(),
		Name:   newUserModel.Name(),
		Email:  newUserModel.Email(),
	}

	return output, nil
}

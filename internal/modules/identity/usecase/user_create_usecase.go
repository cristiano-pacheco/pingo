package usecase

import (
	"context"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

type UserCreateInput struct {
	FirstName string `validate:"required,min=3,max=255"`
	LastName  string `validate:"required,min=3,max=255"`
	Email     string `validate:"required,email"`
	Password  string `validate:"required,min=8"`
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
	logger                       logger.Logger
}

func NewUserCreateUseCase(
	sendEmailConfirmationService service.SendEmailConfirmationService,
	hashService service.HashService,
	userRepo repository.UserRepository,
	validate validator.Validate,
	logger logger.Logger,
) *UserCreateUseCase {
	return &UserCreateUseCase{
		sendEmailConfirmationService,
		hashService,
		userRepo,
		validate,
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

	user, err := uc.userRepository.FindByEmail(ctx, input.Email)
	if err != nil {
		uc.logger.Error("error finding user by email", "error", err)
		return output, err
	}

	if user.ID != 0 {
		return output, errs.ErrEmailAlreadyInUse
	}

	token, err := uc.hashService.GenerateRandomBytes()
	if err != nil {
		message := "error generating random bytes"
		uc.logger.Error(message, "error", err)
		return output, err
	}

	userModel := model.UserModel{
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		Email:             input.Email,
		ConfirmationToken: token,
	}

	createdUser, err := uc.userRepository.Create(ctx, userModel)
	if err != nil {
		message := "error creating user"
		uc.logger.Error(message, "error", err)
		return output, err
	}

	err = uc.sendEmailConfirmationService.Execute(ctx, createdUser.ID)
	if err != nil {
		message := "error sending account confirmation email"
		uc.logger.Error(message, "error", err)
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

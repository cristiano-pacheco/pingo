package usecase

import (
	"context"
	"errors"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/producer"
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
	Email     string `validate:"required,email,max=255"`
}

type UserCreateOutput struct {
	FirstName string
	LastName  string
	Email     string
	UserID    uint64
}

type UserCreateUseCase struct {
	passwordValidator   identity_validator.PasswordValidator
	userCreatedProducer producer.UserCreatedProducer
	userRepository      repository.UserRepository
	hashService         service.HashService
	validate            validator.Validate
	logger              logger.Logger
	otel                otel.Otel
}

func NewUserCreateUseCase(
	passwordValidator identity_validator.PasswordValidator,
	hashService service.HashService,
	userRepository repository.UserRepository,
	userCreatedProducer producer.UserCreatedProducer,
	validate validator.Validate,
	logger logger.Logger,
	otel otel.Otel,
) *UserCreateUseCase {
	return &UserCreateUseCase{
		userCreatedProducer: userCreatedProducer,
		passwordValidator:   passwordValidator,
		userRepository:      userRepository,
		hashService:         hashService,
		validate:            validate,
		logger:              logger,
		otel:                otel,
	}
}

func (uc *UserCreateUseCase) Execute(ctx context.Context, input UserCreateInput) (UserCreateOutput, error) {
	ctx, span := uc.otel.StartSpan(ctx, "UserCreateUseCase.Execute")
	defer span.End()

	output := UserCreateOutput{}

	err := uc.validate.Struct(input)
	if err != nil {
		return output, err
	}

	if err = uc.passwordValidator.Validate(input.Password); err != nil {
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

	passwordHash, err := uc.hashService.GenerateFromPassword([]byte(input.Password))
	if err != nil {
		uc.logger.Error().Msgf("error generating password hash: %v", err)
		return output, err
	}

	pendingUserStatus := enum.UserStatusPending
	userModel := model.UserModel{
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PasswordHash: passwordHash,
		Status:       pendingUserStatus,
	}

	createdUser, err := uc.userRepository.Create(ctx, userModel)
	if err != nil {
		uc.logger.Error().Msgf("error creating user: %v", err)
		return output, err
	}

	message := event.UserCreatedMessage{UserID: createdUser.ID}
	err = uc.userCreatedProducer.Produce(ctx, message)
	if err != nil {
		uc.logger.Error().Msgf("error producing user created event: %v", err)
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

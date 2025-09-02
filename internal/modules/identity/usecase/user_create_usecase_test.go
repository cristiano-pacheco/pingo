package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event"
	producer_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/event/producer/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	validator_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/validator/mocks"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	shared_validator_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/validator/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UserCreateUseCaseTestSuite struct {
	suite.Suite
	sut                     *usecase.UserCreateUseCase
	userCreatedProducerMock *producer_mocks.MockUserCreatedProducer
	passwordValidatorMock   *validator_mocks.MockPasswordValidator
	userRepositoryMock      *repository_mocks.MockUserRepository
	hashServiceMock         *service_mocks.MockHashService
	validatorMock           *shared_validator_mocks.MockValidate
	logger                  logger.Logger
	cfg                     config.Config
	otel                    otel.Otel
}

func (s *UserCreateUseCaseTestSuite) SetupTest() {
	s.userCreatedProducerMock = producer_mocks.NewMockUserCreatedProducer(s.T())
	s.passwordValidatorMock = validator_mocks.NewMockPasswordValidator(s.T())
	s.userRepositoryMock = repository_mocks.NewMockUserRepository(s.T())
	s.hashServiceMock = service_mocks.NewMockHashService(s.T())
	s.validatorMock = shared_validator_mocks.NewMockValidate(s.T())

	s.cfg = config.Config{
		MAIL: config.MAIL{
			Sender: "test@example.com",
		},
		App: config.App{
			BaseURL: "https://example.com",
			Name:    "Test App",
			Version: "1.0.0",
		},
		OpenTelemetry: config.OpenTelemetry{
			Enabled: false,
		},
		Log: config.Log{
			LogLevel: "disabled",
		},
	}

	// Create a simple no-op otel implementation for testing
	s.otel = otel.NewNoopOtel()
	s.logger = logger.New(s.cfg)

	s.sut = usecase.NewUserCreateUseCase(
		s.passwordValidatorMock,
		s.hashServiceMock,
		s.userRepositoryMock,
		s.userCreatedProducerMock,
		s.validatorMock,
		s.logger,
		s.otel,
	)
}

func TestUserCreateUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UserCreateUseCaseTestSuite))
}

func (s *UserCreateUseCaseTestSuite) TestExecute_ValidInput_CreatesUserSuccessfully() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "StrongPassword123!",
	}

	existingUser := model.UserModel{}
	passwordHash := []byte("hashed-password")

	createdUser := model.UserModel{
		ID:           1,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PasswordHash: passwordHash,
		Status:       enum.UserStatusPending,
	}

	expectedEventMessage := event.UserCreatedMessage{UserID: createdUser.ID}

	s.validatorMock.On("Struct", input).Return(nil)
	s.passwordValidatorMock.On("Validate", input.Password).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).
		Return(existingUser, shared_errs.ErrRecordNotFound)
	s.hashServiceMock.On("GenerateFromPassword", []byte(input.Password)).Return(passwordHash, nil)
	s.userRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.UserModel")).Return(createdUser, nil)
	s.userCreatedProducerMock.On("Produce", mock.Anything, expectedEventMessage).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(createdUser.ID, output.UserID)
	s.Equal(input.FirstName, output.FirstName)
	s.Equal(input.LastName, output.LastName)
	s.Equal(input.Email, output.Email)
}

func (s *UserCreateUseCaseTestSuite) TestExecute_ValidationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "",
		LastName:  "",
		Email:     "invalid-email",
		Password:  "weak",
	}
	validationError := errors.New("validation error")

	s.validatorMock.On("Struct", input).Return(validationError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(validationError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *UserCreateUseCaseTestSuite) TestExecute_PasswordValidationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "weakpassword",
	}
	passwordValidationError := errs.ErrPasswordNoUppercase

	s.validatorMock.On("Struct", input).Return(nil)
	s.passwordValidatorMock.On("Validate", input.Password).Return(passwordValidationError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordNoUppercase)
	s.Equal(uint64(0), output.UserID)
}

func (s *UserCreateUseCaseTestSuite) TestExecute_EmailAlreadyInUse_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "StrongPassword123!",
	}

	existingUser := model.UserModel{
		ID:    1,
		Email: input.Email,
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.passwordValidatorMock.On("Validate", input.Password).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(existingUser, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrEmailAlreadyInUse)
	s.Equal(uint64(0), output.UserID)
}

func (s *UserCreateUseCaseTestSuite) TestExecute_FindByEmailRepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "StrongPassword123!",
	}
	repositoryError := errors.New("database error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.passwordValidatorMock.On("Validate", input.Password).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(model.UserModel{}, repositoryError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(repositoryError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *UserCreateUseCaseTestSuite) TestExecute_GenerateFromPasswordFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "StrongPassword123!",
	}
	existingUser := model.UserModel{}
	passwordHashError := errors.New("password hash error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.passwordValidatorMock.On("Validate", input.Password).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).
		Return(existingUser, shared_errs.ErrRecordNotFound)
	s.hashServiceMock.On("GenerateFromPassword", []byte(input.Password)).Return(nil, passwordHashError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(passwordHashError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *UserCreateUseCaseTestSuite) TestExecute_UserRepositoryCreateFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "StrongPassword123!",
	}
	existingUser := model.UserModel{}
	passwordHash := []byte("hashed-password")
	createError := errors.New("create user error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.passwordValidatorMock.On("Validate", input.Password).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).
		Return(existingUser, shared_errs.ErrRecordNotFound)
	s.hashServiceMock.On("GenerateFromPassword", []byte(input.Password)).Return(passwordHash, nil)
	s.userRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.UserModel")).
		Return(model.UserModel{}, createError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(createError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *UserCreateUseCaseTestSuite) TestExecute_UserCreatedProducerFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserCreateInput{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "StrongPassword123!",
	}
	existingUser := model.UserModel{}
	passwordHash := []byte("hashed-password")
	eventError := errors.New("event production error")

	createdUser := model.UserModel{
		ID:           1,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PasswordHash: passwordHash,
		Status:       enum.UserStatusPending,
	}

	expectedEventMessage := event.UserCreatedMessage{UserID: createdUser.ID}

	s.validatorMock.On("Struct", input).Return(nil)
	s.passwordValidatorMock.On("Validate", input.Password).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).
		Return(existingUser, shared_errs.ErrRecordNotFound)
	s.hashServiceMock.On("GenerateFromPassword", []byte(input.Password)).Return(passwordHash, nil)
	s.userRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.UserModel")).Return(createdUser, nil)
	s.userCreatedProducerMock.On("Produce", mock.Anything, expectedEventMessage).Return(eventError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(eventError, err)
	s.Equal(uint64(0), output.UserID)
}

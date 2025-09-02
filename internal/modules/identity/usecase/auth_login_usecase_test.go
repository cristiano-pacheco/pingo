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
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	shared_validator_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/validator/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthLoginUseCaseTestSuite struct {
	suite.Suite
	sut                           *usecase.AuthLoginUseCase
	userAuthenticatedProducerMock *producer_mocks.MockUserAuthenticatedProducer
	userRepositoryMock            *repository_mocks.MockUserRepository
	hashServiceMock               *service_mocks.MockHashService
	validatorMock                 *shared_validator_mocks.MockValidate
	logger                        logger.Logger
	cfg                           config.Config
	otel                          otel.Otel
}

func (s *AuthLoginUseCaseTestSuite) SetupTest() {
	s.cfg = config.Config{
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

	s.otel = otel.NewNoopOtel()
	s.logger = logger.New(s.cfg)

	s.userAuthenticatedProducerMock = producer_mocks.NewMockUserAuthenticatedProducer(s.T())
	s.userRepositoryMock = repository_mocks.NewMockUserRepository(s.T())
	s.hashServiceMock = service_mocks.NewMockHashService(s.T())
	s.validatorMock = shared_validator_mocks.NewMockValidate(s.T())

	s.sut = usecase.NewAuthLoginUseCase(
		s.userAuthenticatedProducerMock,
		s.userRepositoryMock,
		s.validatorMock,
		s.hashServiceMock,
		s.logger,
		s.otel,
	)
}

func TestAuthLoginUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(AuthLoginUseCaseTestSuite))
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_ValidCredentials_ReturnsSuccess() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := model.UserModel{
		ID:           uint64(123),
		Email:        input.Email,
		PasswordHash: []byte("hashed-password"),
		Status:       enum.UserStatusActive,
	}

	message := event.UserAuthenticatedMessage{UserID: user.ID}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.userAuthenticatedProducerMock.On("Produce", mock.Anything, message).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(user.ID, output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_ValidationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "",
		Password: "",
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

func (s *AuthLoginUseCaseTestSuite) TestExecute_UserRepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}
	repositoryError := errors.New("database error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(model.UserModel{}, repositoryError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(repositoryError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_UserNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := model.UserModel{}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, shared_errs.ErrRecordNotFound)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(errs.ErrInvalidCredentials, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_UserNotActive_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := model.UserModel{
		ID:           uint64(123),
		Email:        input.Email,
		PasswordHash: []byte("hashed-password"),
		Status:       enum.UserStatusInactive,
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(errs.ErrUserIsNotActive, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_InvalidPassword_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	user := model.UserModel{
		ID:           uint64(123),
		Email:        input.Email,
		PasswordHash: []byte("hashed-password"),
		Status:       enum.UserStatusActive,
	}

	hashCompareError := errors.New("hash comparison failed")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).
		Return(hashCompareError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(errs.ErrInvalidCredentials, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_ProducerFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := model.UserModel{
		ID:           uint64(123),
		Email:        input.Email,
		PasswordHash: []byte("hashed-password"),
		Status:       enum.UserStatusActive,
	}

	message := event.UserAuthenticatedMessage{UserID: user.ID}
	producerError := errors.New("producer error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.userAuthenticatedProducerMock.On("Produce", mock.Anything, message).Return(producerError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(producerError, err)
	s.Equal(uint64(0), output.UserID)
}

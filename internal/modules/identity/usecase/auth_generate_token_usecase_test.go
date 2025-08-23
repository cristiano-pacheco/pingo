package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	validator_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/validator/mocks"
	"github.com/samber/lo"
)

type AuthGenerateTokenUseCaseTestSuite struct {
	suite.Suite
	sut                            *usecase.AuthGenerateTokenUseCase
	verificationCodeRepositoryMock *repository_mocks.MockVerificationCodeRepository
	userRepositoryMock             *repository_mocks.MockUserRepository
	validatorMock                  *validator_mocks.MockValidate
	tokenServiceMock               *service_mocks.MockTokenService
	logger                         logger.Logger
	cfg                            config.Config
}

func (s *AuthGenerateTokenUseCaseTestSuite) SetupTest() {
	s.cfg = config.Config{
		App: config.App{
			BaseURL: "https://example.com",
			Name:    "Test App",
			Version: "1.0.0",
		},
		Telemetry: config.Telemetry{
			Enabled: false,
		},
		Log: config.Log{
			LogLevel: "disabled",
		},
	}

	otel.Init(s.cfg)
	s.logger = logger.New(s.cfg)

	s.verificationCodeRepositoryMock = repository_mocks.NewMockVerificationCodeRepository(s.T())
	s.userRepositoryMock = repository_mocks.NewMockUserRepository(s.T())
	s.validatorMock = validator_mocks.NewMockValidate(s.T())
	s.tokenServiceMock = service_mocks.NewMockTokenService(s.T())

	s.sut = usecase.NewAuthGenerateTokenUseCase(
		s.verificationCodeRepositoryMock,
		s.userRepositoryMock,
		s.validatorMock,
		s.tokenServiceMock,
		s.logger,
	)
}

func TestAuthGenerateTokenUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(AuthGenerateTokenUseCaseTestSuite))
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_ValidInput_ReturnsToken() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"
	token := "jwt-token"

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
		Email:  "test@example.com",
	}

	verificationCode := model.VerificationCodeModel{
		ID:        1,
		UserID:    userID,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Minute * 5),
		CreatedAt: time.Now(),
		UsedAt:    nil,
	}

	verificationCodeUpdated := verificationCode
	verificationCodeUpdated.UsedAt = lo.ToPtr(time.Now().UTC())

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.verificationCodeRepositoryMock.On("FindByUserAndCode", mock.Anything, userID, code).
		Return(verificationCode, nil)
	s.verificationCodeRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.VerificationCodeModel")).
		Return(nil)
	s.tokenServiceMock.On("GenerateJWT", mock.Anything, user).Return(token, nil)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(token, result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_ValidationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	validationError := errors.New("validation error")

	input := usecase.GenerateTokenInput{
		UserID: 0,
		Code:   "",
	}

	s.validatorMock.On("Struct", input).Return(validationError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(validationError, err)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_UserNotFound_ReturnsInvalidCredentialsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).
		Return(model.UserModel{}, shared_errs.ErrRecordNotFound)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_UserRepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"
	repositoryError := errors.New("database error")

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(model.UserModel{}, repositoryError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(repositoryError, err)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_UserNotActive_ReturnsUserNotActiveError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
		Email:  "test@example.com",
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserIsNotActive)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_VerificationCodeNotFound_ReturnsInvalidCredentialsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
		Email:  "test@example.com",
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.verificationCodeRepositoryMock.On("FindByUserAndCode", mock.Anything, userID, code).
		Return(model.VerificationCodeModel{}, shared_errs.ErrRecordNotFound)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_VerificationCodeRepositoryError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"
	repositoryError := errors.New("database error")

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
		Email:  "test@example.com",
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.verificationCodeRepositoryMock.On("FindByUserAndCode", mock.Anything, userID, code).
		Return(model.VerificationCodeModel{}, repositoryError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(repositoryError, err)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_VerificationCodeUpdateFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"
	updateError := errors.New("update error")

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
		Email:  "test@example.com",
	}

	verificationCode := model.VerificationCodeModel{
		ID:        1,
		UserID:    userID,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Minute * 5),
		CreatedAt: time.Now(),
		UsedAt:    nil,
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.verificationCodeRepositoryMock.On("FindByUserAndCode", mock.Anything, userID, code).
		Return(verificationCode, nil)
	s.verificationCodeRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.VerificationCodeModel")).
		Return(updateError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(updateError, err)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_TokenGenerationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"
	tokenError := errors.New("token generation error")

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
		Email:  "test@example.com",
	}

	verificationCode := model.VerificationCodeModel{
		ID:        1,
		UserID:    userID,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Minute * 5),
		CreatedAt: time.Now(),
		UsedAt:    nil,
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.verificationCodeRepositoryMock.On("FindByUserAndCode", mock.Anything, userID, code).
		Return(verificationCode, nil)
	s.verificationCodeRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.VerificationCodeModel")).
		Return(nil)
	s.tokenServiceMock.On("GenerateJWT", mock.Anything, user).Return("", tokenError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(tokenError, err)
	s.Empty(result.Token)
}

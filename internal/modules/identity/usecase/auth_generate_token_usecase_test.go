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
	validator_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/validator/mocks"
)

type AuthGenerateTokenUseCaseTestSuite struct {
	suite.Suite
	sut                        *usecase.AuthGenerateTokenUseCase
	oneTimeTokenRepositoryMock *repository_mocks.MockOneTimeTokenRepositoryI
	userRepositoryMock         *repository_mocks.MockUserRepositoryI
	validatorMock              *validator_mocks.MockValidate
	tokenServiceMock           *service_mocks.MockTokenServiceI
	hashServiceMock            *service_mocks.MockHashServiceI
	logger                     logger.Logger
	cfg                        config.Config
}

func (s *AuthGenerateTokenUseCaseTestSuite) SetupTest() {
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

	s.logger = logger.New(s.cfg)

	s.oneTimeTokenRepositoryMock = repository_mocks.NewMockOneTimeTokenRepositoryI(s.T())
	s.userRepositoryMock = repository_mocks.NewMockUserRepositoryI(s.T())
	s.validatorMock = validator_mocks.NewMockValidate(s.T())
	s.tokenServiceMock = service_mocks.NewMockTokenServiceI(s.T())
	s.hashServiceMock = service_mocks.NewMockHashServiceI(s.T())

	s.sut = usecase.NewAuthGenerateTokenUseCase(
		s.oneTimeTokenRepositoryMock,
		s.userRepositoryMock,
		s.tokenServiceMock,
		s.hashServiceMock,
		s.validatorMock,
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
	hashedCode := []byte("hashed-code")

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
		Email:  "test@example.com",
	}

	loginVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: hashedCode,
		TokenType: enum.TokenTypeLoginVerification,
		ExpiresAt: time.Now().Add(time.Minute * 5),
		CreatedAt: time.Now(),
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, loginVerificationType).
		Return(oneTimeToken, nil)
	s.hashServiceMock.On("CompareHashAndPassword", hashedCode, []byte(code)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, userID, loginVerificationType).
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

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_OneTimeTokenNotFound_ReturnsInvalidCredentialsError() {
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

	loginVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, loginVerificationType).
		Return(model.OneTimeTokenModel{}, shared_errs.ErrRecordNotFound)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_OneTimeTokenRepositoryError_ReturnsError() {
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

	loginVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, loginVerificationType).
		Return(model.OneTimeTokenModel{}, repositoryError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(repositoryError, err)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_HashComparisonFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"
	hashedCode := []byte("hashed-code")
	hashError := errors.New("hash comparison failed")

	input := usecase.GenerateTokenInput{
		UserID: userID,
		Code:   code,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
		Email:  "test@example.com",
	}

	loginVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: hashedCode,
		TokenType: enum.TokenTypeLoginVerification,
		ExpiresAt: time.Now().Add(time.Minute * 5),
		CreatedAt: time.Now(),
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, loginVerificationType).
		Return(oneTimeToken, nil)
	s.hashServiceMock.On("CompareHashAndPassword", hashedCode, []byte(code)).Return(hashError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(hashError, err)
	s.Empty(result.Token)
}

func (s *AuthGenerateTokenUseCaseTestSuite) TestExecute_TokenGenerationFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	code := "123456"
	hashedCode := []byte("hashed-code")
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

	loginVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: hashedCode,
		TokenType: enum.TokenTypeLoginVerification,
		ExpiresAt: time.Now().Add(time.Minute * 5),
		CreatedAt: time.Now(),
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, loginVerificationType).
		Return(oneTimeToken, nil)
	s.hashServiceMock.On("CompareHashAndPassword", hashedCode, []byte(code)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, userID, loginVerificationType).
		Return(nil)
	s.tokenServiceMock.On("GenerateJWT", mock.Anything, user).Return("", tokenError)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(tokenError, err)
	s.Empty(result.Token)
}

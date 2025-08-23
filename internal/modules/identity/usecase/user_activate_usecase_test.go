package usecase_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	validator_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/validator/mocks"
)

type UserActivateUseCaseTestSuite struct {
	suite.Suite
	sut                        *usecase.UserActivateUseCase
	oneTimeTokenRepositoryMock *repository_mocks.MockOneTimeTokenRepository
	userRepositoryMock         *repository_mocks.MockUserRepository
	validateMock               *validator_mocks.MockValidate
	logger                     logger.Logger
	cfg                        config.Config
}

func (s *UserActivateUseCaseTestSuite) SetupTest() {
	s.cfg = config.Config{
		MAIL: config.MAIL{
			Sender: "test@example.com",
		},
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

	s.oneTimeTokenRepositoryMock = repository_mocks.NewMockOneTimeTokenRepository(s.T())
	s.userRepositoryMock = repository_mocks.NewMockUserRepository(s.T())
	s.validateMock = validator_mocks.NewMockValidate(s.T())
	s.logger = logger.New(s.cfg)

	s.sut = usecase.NewUserActivateUseCase(
		s.oneTimeTokenRepositoryMock,
		s.userRepositoryMock,
		s.validateMock,
		s.logger,
	)
}

func TestUserActivateUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UserActivateUseCaseTestSuite))
}

func (s *UserActivateUseCaseTestSuite) TestExecute_ValidInput_ReturnsSuccess() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	confirmationTokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: []byte(token),
		TokenType: confirmationTokenType.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, confirmationTokenType).Return(oneTimeToken, nil)
	s.userRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.UserModel")).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, userID, confirmationTokenType).Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.NoError(err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_ValidationError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserActivateInput{
		UserID: 0,
		Token:  "",
	}

	validationError := errors.New("validation error")
	s.validateMock.On("Struct", input).Return(validationError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(validationError, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_InvalidTokenBase64_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.UserActivateInput{
		UserID: uint64(123),
		Token:  "invalid-base64!@#",
	}

	s.validateMock.On("Struct", input).Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_UserNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	dbError := errors.New("database error")
	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(model.UserModel{}, dbError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(dbError, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_UserNotInPendingStatus_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusActive,
	}

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(errs.ErrUserNotInPendingStatus, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_OneTimeTokenNotFound_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	confirmationTokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	tokenError := errors.New("token not found")

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, confirmationTokenType).Return(model.OneTimeTokenModel{}, tokenError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(tokenError, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_InvalidTokenHash_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	confirmationTokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: []byte("different-token"),
		TokenType: confirmationTokenType.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, confirmationTokenType).Return(oneTimeToken, nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(errs.ErrInvalidAccountConfirmationToken, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_EmptyTokenID_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	confirmationTokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        0,
		UserID:    userID,
		TokenHash: []byte(token),
		TokenType: confirmationTokenType.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, confirmationTokenType).Return(oneTimeToken, nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(errs.ErrInvalidAccountConfirmationToken, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_UserUpdateError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	confirmationTokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: []byte(token),
		TokenType: confirmationTokenType.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	updateError := errors.New("update error")

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, confirmationTokenType).Return(oneTimeToken, nil)
	s.userRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.UserModel")).Return(updateError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(updateError, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_TokenDeleteError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	confirmationTokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: []byte(token),
		TokenType: confirmationTokenType.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	deleteError := errors.New("delete error")

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, confirmationTokenType).Return(oneTimeToken, nil)
	s.userRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.UserModel")).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, userID, confirmationTokenType).Return(deleteError)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Error(err)
	s.Equal(deleteError, err)
}

func (s *UserActivateUseCaseTestSuite) TestExecute_UserFoundButRecordNotFoundError_ContinuesExecution() {
	// Arrange
	ctx := context.Background()
	userID := uint64(123)
	token := "valid-token"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	input := usecase.UserActivateInput{
		UserID: userID,
		Token:  encodedToken,
	}

	user := model.UserModel{
		ID:     userID,
		Status: enum.UserStatusPending,
	}

	confirmationTokenType, _ := enum.NewTokenTypeEnum(enum.TokenTypeAccountConfirmation)
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    userID,
		TokenHash: []byte(token),
		TokenType: confirmationTokenType.String(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	s.validateMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, userID).Return(user, shared_errs.ErrRecordNotFound)
	s.oneTimeTokenRepositoryMock.On("Find", mock.Anything, userID, confirmationTokenType).Return(oneTimeToken, nil)
	s.userRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.UserModel")).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, userID, confirmationTokenType).Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.NoError(err)
}

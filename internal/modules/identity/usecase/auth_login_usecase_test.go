package usecase_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	repository_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	logger_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/logger/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	validator_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/validator/mocks"
)

type AuthLoginUseCaseTestSuite struct {
	suite.Suite
	sut                              *usecase.AuthLoginUseCase
	oneTimeTokenRepositoryMock       *repository_mocks.MockOneTimeTokenRepository
	userRepositoryMock               *repository_mocks.MockUserRepository
	validatorMock                    *validator_mocks.MockValidate
	hashServiceMock                  *service_mocks.MockHashService
	sendEmailVerificationServiceMock *service_mocks.MockSendEmailVerificationCodeService
	loggerMock                       *logger_mocks.MockLogger
	cfg                              config.Config
	otel                             otel.Otel
}

func (s *AuthLoginUseCaseTestSuite) SetupTest() {
	s.oneTimeTokenRepositoryMock = repository_mocks.NewMockOneTimeTokenRepository(s.T())
	s.userRepositoryMock = repository_mocks.NewMockUserRepository(s.T())
	s.validatorMock = validator_mocks.NewMockValidate(s.T())
	s.hashServiceMock = service_mocks.NewMockHashService(s.T())
	s.sendEmailVerificationServiceMock = service_mocks.NewMockSendEmailVerificationCodeService(s.T())
	s.loggerMock = logger_mocks.NewMockLogger(s.T())

	s.cfg = config.Config{
		App: config.App{
			Name:    "Test App",
			Version: "1.0.0",
		},
		OpenTelemetry: config.OpenTelemetry{
			Enabled: false,
		},
	}

	// Create a simple no-op otel implementation for testing
	s.otel = otel.NewNoopOtel()

	s.sut = usecase.NewAuthLoginUseCase(
		s.oneTimeTokenRepositoryMock,
		s.userRepositoryMock,
		s.validatorMock,
		s.hashServiceMock,
		s.sendEmailVerificationServiceMock,
		s.loggerMock,
		s.otel,
	)
}

func TestAuthLoginUseCaseSuite(t *testing.T) {
	suite.Run(t, new(AuthLoginUseCaseTestSuite))
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_ValidCredentials_ReturnsUserID() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    user.ID,
		TokenHash: []byte("hashedtoken"),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return([]byte("hashedtoken"), nil)
	s.oneTimeTokenRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(oneTimeToken, nil)
	s.sendEmailVerificationServiceMock.On("Execute", mock.Anything, mock.AnythingOfType("service.SendEmailVerificationCodeInput")).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(user.ID, output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_InvalidInput_ReturnsValidationError() {
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

func (s *AuthLoginUseCaseTestSuite) TestExecute_UserNotFound_ReturnsInvalidCredentials() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "notfound@example.com",
		Password: "password",
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).
		Return(model.UserModel{}, shared_errs.ErrRecordNotFound)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_UserNotActive_ReturnsUserNotActiveError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "inactive",
		PasswordHash: []byte("hashedpassword"),
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserIsNotActive)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_InvalidPassword_ReturnsInvalidCredentials() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}
	passwordError := errors.New("password mismatch")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(passwordError)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_FindByEmailDatabaseError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	dbError := errors.New("database connection error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(model.UserModel{}, dbError)
	logger := zerolog.New(os.Stderr)
	s.loggerMock.On("Error").Return(logger.Error())

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(dbError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_DeleteVerificationCodeError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}
	deleteError := errors.New("delete verification code error")

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(deleteError)
	logger := zerolog.New(os.Stderr)
	s.loggerMock.On("Error").Return(logger.Error())

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(deleteError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_CreateVerificationCodeError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}
	createError := errors.New("create verification code error")

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return([]byte("hashedtoken"), nil)
	s.oneTimeTokenRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(model.OneTimeTokenModel{}, createError)
	logger := zerolog.New(os.Stderr)
	s.loggerMock.On("Error").Return(logger.Error())

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(createError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_HashServiceError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}
	hashError := errors.New("hash service error")

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return(nil, hashError)
	logger := zerolog.New(os.Stderr)
	s.loggerMock.On("Error").Return(logger.Error())

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(hashError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_SendEmailError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    user.ID,
		TokenHash: []byte("hashedtoken"),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}
	emailError := errors.New("send email error")

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return([]byte("hashedtoken"), nil)
	s.oneTimeTokenRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(oneTimeToken, nil)
	s.sendEmailVerificationServiceMock.On("Execute", mock.Anything, mock.AnythingOfType("service.SendEmailVerificationCodeInput")).
		Return(emailError)
	logger := zerolog.New(os.Stderr)
	s.loggerMock.On("Error").Return(logger.Error())

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(emailError, err)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_DeleteVerificationCodeNotFound_ContinuesExecution() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}
	oneTimeToken := model.OneTimeTokenModel{
		ID:        1,
		UserID:    user.ID,
		TokenHash: []byte("hashedtoken"),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).
		Return(shared_errs.ErrRecordNotFound)
	s.hashServiceMock.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return([]byte("hashedtoken"), nil)
	s.oneTimeTokenRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.OneTimeTokenModel")).
		Return(oneTimeToken, nil)
	s.sendEmailVerificationServiceMock.On("Execute", mock.Anything, mock.AnythingOfType("service.SendEmailVerificationCodeInput")).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(user.ID, output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_EmptyUser_ReturnsInvalidCredentials() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	emptyUser := model.UserModel{} // Empty user with ID = 0

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(emptyUser, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Equal(uint64(0), output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_ValidCredentialsGeneratesVerificationCode() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return([]byte("hashedtoken"), nil)
	s.oneTimeTokenRepositoryMock.On("Create", mock.Anything, mock.MatchedBy(func(ott model.OneTimeTokenModel) bool {
		return ott.UserID == user.ID && len(ott.TokenHash) > 0
	})).Return(model.OneTimeTokenModel{
		ID:        1,
		UserID:    user.ID,
		TokenHash: []byte("hashedtoken"),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}, nil)
	s.sendEmailVerificationServiceMock.On("Execute", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
		if sei, ok := input.(service.SendEmailVerificationCodeInput); ok {
			return sei.UserID == user.ID && sei.Code != ""
		}
		return false
	})).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(user.ID, output.UserID)
}

func (s *AuthLoginUseCaseTestSuite) TestExecute_VerificationCodeFormat_SixDigitCode() {
	// Arrange
	ctx := context.Background()
	input := usecase.AuthLoginInput{
		Email:    "test@example.com",
		Password: "validpassword",
	}
	user := model.UserModel{
		ID:           1,
		Email:        "test@example.com",
		Status:       "active",
		PasswordHash: []byte("hashedpassword"),
	}

	emailVerificationType, _ := enum.NewTokenTypeEnum(enum.TokenTypeLoginVerification)
	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.oneTimeTokenRepositoryMock.On("Delete", mock.Anything, user.ID, emailVerificationType).Return(nil)
	s.hashServiceMock.On("GenerateFromPassword", mock.AnythingOfType("[]uint8")).Return([]byte("hashedtoken"), nil)

	// Capture the created one-time token to verify its properties
	var capturedCode string
	s.oneTimeTokenRepositoryMock.On("Create", mock.Anything, mock.MatchedBy(func(ott model.OneTimeTokenModel) bool {
		return ott.UserID == user.ID && len(ott.TokenHash) > 0
	})).Return(model.OneTimeTokenModel{
		ID:        1,
		UserID:    user.ID,
		TokenHash: []byte("hashedtoken"),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}, nil)

	s.sendEmailVerificationServiceMock.On("Execute", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
		if sei, ok := input.(service.SendEmailVerificationCodeInput); ok {
			capturedCode = sei.Code
			return sei.UserID == user.ID && len(sei.Code) == 6
		}
		return false
	})).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(user.ID, output.UserID)
	s.Len(capturedCode, 6)
	s.Regexp(`^\d{6}$`, capturedCode) // Verify it's 6 digits
}

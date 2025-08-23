package usecase_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

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
	verificationCodeRepositoryMock   *repository_mocks.MockVerificationCodeRepository
	userRepositoryMock               *repository_mocks.MockUserRepository
	validatorMock                    *validator_mocks.MockValidate
	hashServiceMock                  *service_mocks.MockHashService
	sendEmailVerificationServiceMock *service_mocks.MockSendEmailVerificationCodeService
	loggerMock                       *logger_mocks.MockLogger
	cfg                              config.Config
}

func (s *AuthLoginUseCaseTestSuite) SetupTest() {
	s.verificationCodeRepositoryMock = repository_mocks.NewMockVerificationCodeRepository(s.T())
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
		Telemetry: config.Telemetry{
			Enabled: false,
		},
	}

	otel.Init(s.cfg)

	s.sut = usecase.NewAuthLoginUseCase(
		s.verificationCodeRepositoryMock,
		s.userRepositoryMock,
		s.validatorMock,
		s.hashServiceMock,
		s.sendEmailVerificationServiceMock,
		s.loggerMock,
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
	verificationCode := model.VerificationCodeModel{
		ID:        1,
		UserID:    user.ID,
		Code:      "123456",
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.verificationCodeRepositoryMock.On("DeleteByUserID", mock.Anything, user.ID).Return(nil)
	s.verificationCodeRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.VerificationCodeModel")).
		Return(verificationCode, nil)
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
	validationError := fmt.Errorf("validation error")

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
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(model.UserModel{}, shared_errs.ErrRecordNotFound)

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
	passwordError := fmt.Errorf("password mismatch")

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
	dbError := fmt.Errorf("database connection error")

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
	deleteError := fmt.Errorf("delete verification code error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.verificationCodeRepositoryMock.On("DeleteByUserID", mock.Anything, user.ID).Return(deleteError)
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
	createError := fmt.Errorf("create verification code error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.verificationCodeRepositoryMock.On("DeleteByUserID", mock.Anything, user.ID).Return(nil)
	s.verificationCodeRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.VerificationCodeModel")).
		Return(model.VerificationCodeModel{}, createError)
	logger := zerolog.New(os.Stderr)
	s.loggerMock.On("Error").Return(logger.Error())

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(createError, err)
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
	verificationCode := model.VerificationCodeModel{
		ID:        1,
		UserID:    user.ID,
		Code:      "123456",
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}
	emailError := fmt.Errorf("send email error")

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.verificationCodeRepositoryMock.On("DeleteByUserID", mock.Anything, user.ID).Return(nil)
	s.verificationCodeRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.VerificationCodeModel")).
		Return(verificationCode, nil)
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
	verificationCode := model.VerificationCodeModel{
		ID:        1,
		UserID:    user.ID,
		Code:      "123456",
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.verificationCodeRepositoryMock.On("DeleteByUserID", mock.Anything, user.ID).Return(shared_errs.ErrRecordNotFound)
	s.verificationCodeRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.VerificationCodeModel")).
		Return(verificationCode, nil)
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

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.verificationCodeRepositoryMock.On("DeleteByUserID", mock.Anything, user.ID).Return(nil)
	s.verificationCodeRepositoryMock.On("Create", mock.Anything, mock.MatchedBy(func(vc model.VerificationCodeModel) bool {
		return vc.UserID == user.ID && vc.Code != "" && len(vc.Code) == 6
	})).Return(model.VerificationCodeModel{
		ID:        1,
		UserID:    user.ID,
		Code:      "123456",
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

	s.validatorMock.On("Struct", input).Return(nil)
	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(user, nil)
	s.hashServiceMock.On("CompareHashAndPassword", user.PasswordHash, []byte(input.Password)).Return(nil)
	s.verificationCodeRepositoryMock.On("DeleteByUserID", mock.Anything, user.ID).Return(nil)

	// Capture the created verification code to verify its format
	var capturedCode string
	s.verificationCodeRepositoryMock.On("Create", mock.Anything, mock.MatchedBy(func(vc model.VerificationCodeModel) bool {
		capturedCode = vc.Code
		return vc.UserID == user.ID && len(vc.Code) == 6
	})).Return(model.VerificationCodeModel{
		ID:        1,
		UserID:    user.ID,
		Code:      "123456",
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}, nil)

	s.sendEmailVerificationServiceMock.On("Execute", mock.Anything, mock.MatchedBy(func(input interface{}) bool {
		if sei, ok := input.(service.SendEmailVerificationCodeInput); ok {
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

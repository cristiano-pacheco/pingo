package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	email_template_service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer"
	mailer_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SendEmailVerificationCodeServiceTestSuite struct {
	suite.Suite
	sut                  *service.SendEmailVerificationCodeService
	emailTemplateService *email_template_service_mocks.MockEmailTemplateServiceI
	mailerSMTP           *mailer_mocks.MockSMTP
	userRepository       *mocks.MockUserRepositoryI
	logger               logger.Logger
	cfg                  config.Config
}

func (s *SendEmailVerificationCodeServiceTestSuite) SetupTest() {
	s.emailTemplateService = email_template_service_mocks.NewMockEmailTemplateServiceI(s.T())
	s.mailerSMTP = mailer_mocks.NewMockSMTP(s.T())
	s.userRepository = mocks.NewMockUserRepositoryI(s.T())

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

	// Use real logger but with disabled level
	s.logger = logger.New(s.cfg)

	s.sut = service.NewSendEmailVerificationCodeService(
		s.emailTemplateService,
		s.mailerSMTP,
		s.userRepository,
		s.logger,
		s.cfg,
	)
}

func TestSendEmailVerificationCodeServiceSuite(t *testing.T) {
	suite.Run(t, new(SendEmailVerificationCodeServiceTestSuite))
}

func (s *SendEmailVerificationCodeServiceTestSuite) TestExecute_ValidInput_SendsEmailSuccessfully() {
	// Arrange
	userID := uint64(123)
	verificationCode := "123456"
	user := model.UserModel{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	expectedName := "John Doe"
	expectedContent := "<html>Verification Code: 123456</html>"
	expectedMailData := mailer.MailData{
		Sender:  s.cfg.MAIL.Sender,
		ToName:  expectedName,
		ToEmail: user.Email,
		Subject: "Verification Code",
		Content: expectedContent,
	}

	input := service.SendEmailVerificationCodeInput{
		UserID: userID,
		Code:   verificationCode,
	}

	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.emailTemplateService.
		On("CompileAuthVerificationCodeTemplate", expectedName, verificationCode).
		Return(expectedContent, nil)
	s.mailerSMTP.On("Send", mock.Anything, expectedMailData).Return(nil)

	// Act
	err := s.sut.Execute(context.Background(), input)

	// Assert
	s.Require().NoError(err)
}

func (s *SendEmailVerificationCodeServiceTestSuite) TestExecute_UserNotFound_ReturnsError() {
	// Arrange
	userID := uint64(123)
	verificationCode := "123456"
	userNotFoundError := errors.New("user not found")

	input := service.SendEmailVerificationCodeInput{
		UserID: userID,
		Code:   verificationCode,
	}

	s.userRepository.On("FindByID", mock.Anything, userID).Return(model.UserModel{}, userNotFoundError)

	// Act
	err := s.sut.Execute(context.Background(), input)

	// Assert
	s.Require().Error(err)
	s.Equal(userNotFoundError, err)
}

func (s *SendEmailVerificationCodeServiceTestSuite) TestExecute_EmailTemplateCompilationFails_ReturnsError() {
	// Arrange
	userID := uint64(123)
	verificationCode := "123456"
	user := model.UserModel{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	expectedName := "John Doe"
	templateError := errors.New("template compilation failed")

	input := service.SendEmailVerificationCodeInput{
		UserID: userID,
		Code:   verificationCode,
	}

	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.emailTemplateService.On("CompileAuthVerificationCodeTemplate", expectedName, verificationCode).
		Return("", templateError)

	// Act
	err := s.sut.Execute(context.Background(), input)

	// Assert
	s.Require().Error(err)
	s.Equal(templateError, err)
}

func (s *SendEmailVerificationCodeServiceTestSuite) TestExecute_EmailSendingFails_ReturnsError() {
	// Arrange
	userID := uint64(123)
	verificationCode := "123456"
	user := model.UserModel{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	expectedName := "John Doe"
	expectedContent := "<html>Verification Code: 123456</html>"
	expectedMailData := mailer.MailData{
		Sender:  s.cfg.MAIL.Sender,
		ToName:  expectedName,
		ToEmail: user.Email,
		Subject: "Verification Code",
		Content: expectedContent,
	}
	sendError := errors.New("failed to send email")

	input := service.SendEmailVerificationCodeInput{
		UserID: userID,
		Code:   verificationCode,
	}

	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.emailTemplateService.On("CompileAuthVerificationCodeTemplate", expectedName, verificationCode).
		Return(expectedContent, nil)
	s.mailerSMTP.On("Send", mock.Anything, expectedMailData).Return(sendError)

	// Act
	err := s.sut.Execute(context.Background(), input)

	// Assert
	s.Require().Error(err)
	s.Equal(sendError, err)
}

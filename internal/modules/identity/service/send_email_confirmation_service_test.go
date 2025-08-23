package service_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/mocks"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	email_template_service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer"
	mailer_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SendEmailConfirmationServiceTestSuite struct {
	suite.Suite
	sut                  service.SendEmailConfirmationService
	emailTemplateService *email_template_service_mocks.MockEmailTemplateService
	mailerSMTP           *mailer_mocks.MockSMTP
	userRepository       *mocks.MockUserRepository
	logger               logger.Logger
	cfg                  config.Config
}

func (s *SendEmailConfirmationServiceTestSuite) SetupTest() {
	s.emailTemplateService = email_template_service_mocks.NewMockEmailTemplateService(s.T())
	s.mailerSMTP = mailer_mocks.NewMockSMTP(s.T())
	s.userRepository = mocks.NewMockUserRepository(s.T())

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

	// Initialize otel for testing
	otel.Init(s.cfg)

	// Use real logger but with disabled level
	s.logger = logger.New(s.cfg)

	s.sut = service.NewSendEmailConfirmationService(
		s.emailTemplateService,
		s.mailerSMTP,
		s.userRepository,
		s.logger,
		s.cfg,
	)
}

func TestSendEmailConfirmationServiceSuite(t *testing.T) {
	suite.Run(t, new(SendEmailConfirmationServiceTestSuite))
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_ValidUser_SendsEmailSuccessfully() {
	// Arrange
	userID := uint64(123)
	confirmationToken := []byte("test-token")
	user := model.UserModel{
		ID:                userID,
		FirstName:         "John",
		LastName:          "Doe",
		Email:             "john.doe@example.com",
		ConfirmationToken: confirmationToken,
	}

	encodedToken := base64.StdEncoding.EncodeToString(confirmationToken)
	expectedLink := "https://example.com/user/confirmation?id=123&token=" + encodedToken
	expectedName := "John Doe"

	emailTemplateInput := service.AccountConfirmationInput{
		Name:                    expectedName,
		AccountConfirmationLink: expectedLink,
	}

	expectedContent := "<html>confirmation email content</html>"

	expectedMailData := mailer.MailData{
		Sender:  "test@example.com",
		ToName:  expectedName,
		ToEmail: "john.doe@example.com",
		Subject: "Account Confirmation",
		Content: expectedContent,
	}

	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.emailTemplateService.
		On("CompileAccountConfirmationTemplate", emailTemplateInput).
		Return(expectedContent, nil)
	s.mailerSMTP.On("Send", mock.Anything, expectedMailData).Return(nil)

	// Act
	err := s.sut.Execute(context.Background(), userID)

	// Assert
	s.Require().NoError(err)
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_UserNotFound_ReturnsError() {
	// Arrange
	userID := uint64(123)
	expectedError := errors.New("user not found")

	s.userRepository.On("FindByID", mock.Anything, userID).Return(model.UserModel{}, expectedError)

	// Act
	err := s.sut.Execute(context.Background(), userID)

	// Assert
	s.Require().Error(err)
	s.Equal(expectedError, err)
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_UserWithNilConfirmationToken_ReturnsError() {
	// Arrange
	userID := uint64(123)
	user := model.UserModel{
		ID:                userID,
		FirstName:         "John",
		LastName:          "Doe",
		Email:             "john.doe@example.com",
		ConfirmationToken: nil,
	}

	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)

	// Act
	err := s.sut.Execute(context.Background(), userID)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidAccountConfirmationToken)
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_EmailTemplateCompilationFails_ReturnsError() {
	// Arrange
	userID := uint64(123)
	confirmationToken := []byte("test-token")
	user := model.UserModel{
		ID:                userID,
		FirstName:         "John",
		LastName:          "Doe",
		Email:             "john.doe@example.com",
		ConfirmationToken: confirmationToken,
	}

	encodedToken := base64.StdEncoding.EncodeToString(confirmationToken)
	expectedLink := "https://example.com/user/confirmation?id=123&token=" + encodedToken
	expectedName := "John Doe"

	emailTemplateInput := service.AccountConfirmationInput{
		Name:                    expectedName,
		AccountConfirmationLink: expectedLink,
	}

	templateError := errors.New("template compilation failed")

	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.emailTemplateService.On("CompileAccountConfirmationTemplate", emailTemplateInput).
		Return("", templateError)

	// Act
	err := s.sut.Execute(context.Background(), userID)

	// Assert
	s.Require().Error(err)
	s.Equal(templateError, err)
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_EmailSendingFails_ReturnsError() {
	// Arrange
	userID := uint64(123)
	confirmationToken := []byte("test-token")
	user := model.UserModel{
		ID:                userID,
		FirstName:         "John",
		LastName:          "Doe",
		Email:             "john.doe@example.com",
		ConfirmationToken: confirmationToken,
	}

	encodedToken := base64.StdEncoding.EncodeToString(confirmationToken)
	expectedLink := "https://example.com/user/confirmation?id=123&token=" + encodedToken
	expectedName := "John Doe"

	emailTemplateInput := service.AccountConfirmationInput{
		Name:                    expectedName,
		AccountConfirmationLink: expectedLink,
	}

	expectedContent := "<html>confirmation email content</html>"

	expectedMailData := mailer.MailData{
		Sender:  "test@example.com",
		ToName:  expectedName,
		ToEmail: "john.doe@example.com",
		Subject: "Account Confirmation",
		Content: expectedContent,
	}

	sendError := errors.New("failed to send email")

	s.userRepository.On("FindByID", mock.Anything, userID).Return(user, nil)
	s.emailTemplateService.On("CompileAccountConfirmationTemplate", emailTemplateInput).
		Return(expectedContent, nil)
	s.mailerSMTP.On("Send", mock.Anything, expectedMailData).Return(sendError)

	// Act
	err := s.sut.Execute(context.Background(), userID)

	// Assert
	s.Require().Error(err)
	s.Equal(sendError, err)
}

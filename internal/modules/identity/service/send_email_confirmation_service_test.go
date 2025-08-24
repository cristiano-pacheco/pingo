package service_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	email_template_service_mocks "github.com/cristiano-pacheco/pingo/internal/modules/identity/service/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer"
	mailer_mocks "github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer/mocks"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"
)

// noopOtel implements otel.Otel interface for testing
type noopOtel struct{}

func (n *noopOtel) StartSpan(_ context.Context, name string) (context.Context, trace.Span) {
	return nil, trace.SpanFromContext(nil)
}

type SendEmailConfirmationServiceTestSuite struct {
	suite.Suite
	sut                  service.SendEmailConfirmationService
	emailTemplateService *email_template_service_mocks.MockEmailTemplateService
	mailerSMTP           *mailer_mocks.MockSMTP
	logger               logger.Logger
	cfg                  config.Config
	otel                 otel.Otel
}

func (s *SendEmailConfirmationServiceTestSuite) SetupTest() {
	s.emailTemplateService = email_template_service_mocks.NewMockEmailTemplateService(s.T())
	s.mailerSMTP = mailer_mocks.NewMockSMTP(s.T())

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
	s.otel = &noopOtel{}

	// Use real logger but with disabled level
	s.logger = logger.New(s.cfg)

	s.sut = service.NewSendEmailConfirmationService(
		s.emailTemplateService,
		s.mailerSMTP,
		s.logger,
		s.cfg,
		s.otel,
	)
}

func TestSendEmailConfirmationServiceSuite(t *testing.T) {
	suite.Run(t, new(SendEmailConfirmationServiceTestSuite))
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_ValidUser_SendsEmailSuccessfully() {
	// Arrange
	confirmationToken := []byte("test-token")
	user := model.UserModel{
		ID:        uint64(123),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	input := service.SendEmailConfirmationInput{
		UserModel:             user,
		ConfirmationTokenHash: confirmationToken,
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

	s.emailTemplateService.
		On("CompileAccountConfirmationTemplate", emailTemplateInput).
		Return(expectedContent, nil)
	s.mailerSMTP.On("Send", mock.Anything, expectedMailData).Return(nil)

	// Act
	err := s.sut.Execute(context.Background(), input)

	// Assert
	s.Require().NoError(err)
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_EmailTemplateCompilationFails_ReturnsError() {
	// Arrange
	confirmationToken := []byte("test-token")
	user := model.UserModel{
		ID:        uint64(123),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	input := service.SendEmailConfirmationInput{
		UserModel:             user,
		ConfirmationTokenHash: confirmationToken,
	}

	encodedToken := base64.StdEncoding.EncodeToString(confirmationToken)
	expectedLink := "https://example.com/user/confirmation?id=123&token=" + encodedToken
	expectedName := "John Doe"

	emailTemplateInput := service.AccountConfirmationInput{
		Name:                    expectedName,
		AccountConfirmationLink: expectedLink,
	}

	templateError := errors.New("template compilation failed")

	s.emailTemplateService.On("CompileAccountConfirmationTemplate", emailTemplateInput).
		Return("", templateError)

	// Act
	err := s.sut.Execute(context.Background(), input)

	// Assert
	s.Require().Error(err)
	s.Equal(templateError, err)
}

func (s *SendEmailConfirmationServiceTestSuite) TestExecute_EmailSendingFails_ReturnsError() {
	// Arrange
	confirmationToken := []byte("test-token")
	user := model.UserModel{
		ID:        uint64(123),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	input := service.SendEmailConfirmationInput{
		UserModel:             user,
		ConfirmationTokenHash: confirmationToken,
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

	s.emailTemplateService.On("CompileAccountConfirmationTemplate", emailTemplateInput).
		Return(expectedContent, nil)
	s.mailerSMTP.On("Send", mock.Anything, expectedMailData).Return(sendError)

	// Act
	err := s.sut.Execute(context.Background(), input)

	// Assert
	s.Require().Error(err)
	s.Equal(sendError, err)
}

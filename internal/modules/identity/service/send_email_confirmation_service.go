package service

import (
	"context"
	"fmt"

	"github.com/cristiano-pacheco/pingo/internal/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
)

const sendAccountConfirmationEmailTemplate = "account_confirmation.gohtml"
const sendAccountConfirmationEmailSubject = "Account Confirmation"

type SendEmailConfirmationService interface {
	Execute(ctx context.Context, userID uint64) error
}

type sendEmailConfirmationService struct {
	mailerTemplate mailer.MailerTemplate
	mailer         mailer.SmtpMailer
	userRepository repository.UserRepository
	logger         logger.Logger
	cfg            config.Config
}

func NewSendEmailConfirmationService(
	mailerTemplate mailer.MailerTemplate,
	smtpMailer mailer.SmtpMailer,
	userRepository repository.UserRepository,
	logger logger.Logger,
	cfg config.Config,
) SendEmailConfirmationService {
	return &sendEmailConfirmationService{
		mailerTemplate,
		smtpMailer,
		userRepository,
		logger,
		cfg,
	}
}

func (s *sendEmailConfirmationService) Execute(ctx context.Context, userID uint64) error {
	ctx, span := otel.Trace().StartSpan(ctx, "sendEmailConfirmationService.Execute")
	defer span.End()

	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		message := "error finding user"
		s.logger.Error(message, "error", err)
		return err
	}

	// generate the account confirmation link
	accountConfLink := fmt.Sprintf(
		"%s/user/confirmation?id=%d&token=%s",
		s.cfg.App.BaseURL,
		user.ID(),
		*user.ConfirmationToken(),
	)

	// compile the template
	tplData := struct {
		Name                    string
		AccountConfirmationLink string
	}{
		Name:                    user.Name(),
		AccountConfirmationLink: accountConfLink,
	}

	content, err := s.mailerTemplate.CompileTemplate(sendAccountConfirmationEmailTemplate, tplData)
	if err != nil {
		message := "error compiling template"
		s.logger.Error(message, "error", err)
		return err
	}

	md := mailer.MailData{
		Sender:  s.cfg.MAIL.Sender,
		ToName:  user.Name(),
		ToEmail: user.Email(),
		Subject: sendAccountConfirmationEmailSubject,
		Content: content,
	}

	err = s.mailer.Send(ctx, md)
	if err != nil {
		message := "error sending email"
		s.logger.Error(message, "error", err)
		return err
	}

	return nil
}

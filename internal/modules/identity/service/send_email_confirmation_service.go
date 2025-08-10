package service

import (
	"context"
	"fmt"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
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
	emailTemplateService EmailTemplateService
	mailerSMTP           mailer.SMTP
	userRepository       repository.UserRepository
	logger               logger.Logger
	cfg                  config.Config
}

func NewSendEmailConfirmationService(
	emailTemplateService EmailTemplateService,
	mailerSMTP mailer.SMTP,
	userRepository repository.UserRepository,
	logger logger.Logger,
	cfg config.Config,
) SendEmailConfirmationService {
	return &sendEmailConfirmationService{
		emailTemplateService,
		mailerSMTP,
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

	if user.ConfirmationToken == nil {
		return errs.ErrInvalidAccountConfirmationToken
	}

	confirmationToken := string(user.ConfirmationToken)

	// generate the account confirmation link
	accountConfLink := fmt.Sprintf(
		"%s/user/confirmation?id=%d&token=%s",
		s.cfg.App.BaseURL,
		user.ID,
		confirmationToken,
	)

	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	emailTemplateInput := AccountConfirmationInput{
		Name:                    name,
		AccountConfirmationLink: accountConfLink,
	}
	content, err := s.emailTemplateService.CompileAccountConfirmationTemplate(ctx, emailTemplateInput)
	md := mailer.MailData{
		Sender:  s.cfg.MAIL.Sender,
		ToName:  name,
		ToEmail: user.Email,
		Subject: sendAccountConfirmationEmailSubject,
		Content: content,
	}

	err = s.mailerSMTP.Send(ctx, md)
	if err != nil {
		message := "error sending email"
		s.logger.Error(message, "error", err)
		return err
	}

	return nil
}

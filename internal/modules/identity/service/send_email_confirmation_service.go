package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/cristiano-pacheco/go-otel/trace"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer"
)

const sendAccountConfirmationEmailSubject = "Account Confirmation"

type SendEmailConfirmationInput struct {
	UserModel             model.UserModel
	ConfirmationTokenHash []byte
}

type SendEmailConfirmationServiceI interface {
	Execute(ctx context.Context, input SendEmailConfirmationInput) error
}

type SendEmailConfirmationService struct {
	emailTemplateService EmailTemplateServiceI
	mailerSMTP           mailer.SMTP
	logger               logger.Logger
	cfg                  config.Config
}

var _ SendEmailConfirmationServiceI = (*SendEmailConfirmationService)(nil)

func NewSendEmailConfirmationService(
	emailTemplateService EmailTemplateServiceI,
	mailerSMTP mailer.SMTP,
	logger logger.Logger,
	cfg config.Config,
) *SendEmailConfirmationService {
	return &SendEmailConfirmationService{
		emailTemplateService,
		mailerSMTP,
		logger,
		cfg,
	}
}

func (s *SendEmailConfirmationService) Execute(ctx context.Context, input SendEmailConfirmationInput) error {
	ctx, span := trace.Span(ctx, "SendEmailConfirmationService.Execute")
	defer span.End()

	confirmationToken := base64.StdEncoding.EncodeToString(input.ConfirmationTokenHash)

	// generate the account confirmation link
	accountConfLink := fmt.Sprintf(
		"%s/user/confirmation?id=%d&token=%s",
		s.cfg.App.BaseURL,
		input.UserModel.ID,
		confirmationToken,
	)

	name := fmt.Sprintf("%s %s", input.UserModel.FirstName, input.UserModel.LastName)
	emailTemplateInput := AccountConfirmationInput{
		Name:                    name,
		AccountConfirmationLink: accountConfLink,
	}
	content, err := s.emailTemplateService.CompileAccountConfirmationTemplate(emailTemplateInput)
	if err != nil {
		s.logger.Error().Msgf("error compiling account confirmation template: %v", err)
		return err
	}

	md := mailer.MailData{
		Sender:  s.cfg.MAIL.Sender,
		ToName:  name,
		ToEmail: input.UserModel.Email,
		Subject: sendAccountConfirmationEmailSubject,
		Content: content,
	}

	err = s.mailerSMTP.Send(ctx, md)
	if err != nil {
		s.logger.Error().Msgf("error sending the confirmation email for the user ID %d: %v", input.UserModel.ID, err)
		return err
	}

	return nil
}

package service

import (
	"context"
	"fmt"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/ui/email/templates"
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
	mailerTemplate mailer.Template
	mailerSMTP     mailer.SMTP
	userRepository repository.UserRepository
	logger         logger.Logger
	cfg            config.Config
}

func NewSendEmailConfirmationService(
	mailerTemplate mailer.Template,
	mailerSMTP mailer.SMTP,
	userRepository repository.UserRepository,
	logger logger.Logger,
	cfg config.Config,
) SendEmailConfirmationService {
	return &sendEmailConfirmationService{
		mailerTemplate,
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

	// compile the template
	tplData := struct {
		Name                    string
		AccountConfirmationLink string
	}{
		Name:                    fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		AccountConfirmationLink: accountConfLink,
	}

	compileTemplateInput := mailer.CompileTemplateInput{
		TemplateName:        sendAccountConfirmationEmailTemplate,
		LayoutTpl:           "layout_default.gohtml",
		TemplatePath:        "",
		TemplateSectionName: "htmlBody",
		TemplateFS:          templates.EmailTemplatesFS,
		Data:                tplData,
	}

	content, err := s.mailerTemplate.CompileTemplate(compileTemplateInput)
	if err != nil {
		message := "error compiling template"
		s.logger.Error(message, "error", err)
		return err
	}

	md := mailer.MailData{
		Sender:  s.cfg.MAIL.Sender,
		ToName:  fmt.Sprintf("%s %s", user.FirstName, user.LastName),
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

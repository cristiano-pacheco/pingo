package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
)

const emailVerificationCodeSubject = "Verification Code"

type SendEmailVerificationCodeService interface {
	Execute(ctx context.Context, input SendEmailVerificationCodeInput) error
}

type sendEmailVerificationCodeService struct {
	emailTemplateService EmailTemplateService
	mailerSMTP           mailer.SMTP
	userRepository       repository.UserRepository
	logger               logger.Logger
	cfg                  config.Config
}

func NewSendEmailVerificationCodeService(
	emailTemplateService EmailTemplateService,
	mailerSMTP mailer.SMTP,
	userRepository repository.UserRepository,
	logger logger.Logger,
	cfg config.Config,
) SendEmailVerificationCodeService {
	return &sendEmailVerificationCodeService{
		emailTemplateService,
		mailerSMTP,
		userRepository,
		logger,
		cfg,
	}
}

type SendEmailVerificationCodeInput struct {
	UserID uint64
	Code   string
}

func (s *sendEmailVerificationCodeService) Execute(
	ctx context.Context,
	input SendEmailVerificationCodeInput,
) error {
	ctx, span := otel.Trace().StartSpan(ctx, "SendEmailVerificationCodeService.Execute")
	defer span.End()

	user, err := s.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		s.logger.Error().Msgf("error finding user with ID %d: %v", input.UserID, err)
		return err
	}

	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	// generate 6 digit random numbers
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	content, err := s.emailTemplateService.CompileAuthVerificationCodeTemplate(ctx, name, code)
	md := mailer.MailData{
		Sender:  s.cfg.MAIL.Sender,
		ToName:  name,
		ToEmail: user.Email,
		Subject: emailVerificationCodeSubject,
		Content: content,
	}

	err = s.mailerSMTP.Send(ctx, md)
	if err != nil {
		s.logger.Error().Msgf("error sending verification code email for the user ID %d: %v", user.ID, err)
		return err
	}

	return nil
}

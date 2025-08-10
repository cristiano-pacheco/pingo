package mailer

import (
	"context"

	"github.com/go-mail/mail/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type SMTP interface {
	Send(ctx context.Context, md MailData) error
}

type smtp struct {
	dialer *mail.Dialer
}

func NewSMTP(dialer *mail.Dialer) SMTP {
	return &smtp{dialer}
}

func (m *smtp) Send(ctx context.Context, md MailData) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("method", "SMTP.Send"))
	defer span.End()

	msg := mail.NewMessage()
	msg.SetHeader("To", md.ToEmail)
	msg.SetHeader("From", md.Sender)
	msg.SetHeader("Subject", md.Subject)
	msg.SetBody("text/html", md.Content)

	err := m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}

package mailer

import "go.uber.org/fx"

var Module = fx.Module(
	"kernel/mailer",
	fx.Provide(NewMailerTemplate),
	fx.Provide(NewSMTPMailer),
	fx.Provide(NewDialer),
)

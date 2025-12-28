package mailer

import "go.uber.org/fx"

var Module = fx.Module(
	"sdk/mailer",
	fx.Provide(NewMailerTemplate),
	fx.Provide(NewSMTP),
	fx.Provide(NewDialer),
)

package mailer

import (
	"github.com/go-mail/mail/v2"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
)

func NewDialer(cfg config.Config) *mail.Dialer {
	return mail.NewDialer(cfg.MAIL.Host, cfg.MAIL.Port, cfg.MAIL.Username, cfg.MAIL.Password)
}

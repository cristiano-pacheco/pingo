package validator

import (
	"net/mail"
	"net/url"

	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/errs"
)

type ContactValidatorI interface {
	Validate(contactType, contactData string) error
}

type ContactValidator struct {
}

var _ ContactValidatorI = (*ContactValidator)(nil)

func NewContactValidator() *ContactValidator {
	return &ContactValidator{}
}

func (v *ContactValidator) Validate(contactType, contactData string) error {
	switch contactType {
	case enum.ContactTypeEmail:
		return v.validateEmail(contactData)
	case enum.ContactTypeWebhook:
		return v.validateWebhook(contactData)
	}
	return nil
}

func (v *ContactValidator) validateEmail(contactData string) error {
	_, err := mail.ParseAddress(contactData)
	if err != nil {
		return errs.ErrInvalidContactEmail
	}
	return nil
}

func (v *ContactValidator) validateWebhook(contactData string) error {
	parsedURL, err := url.ParseRequestURI(contactData)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return errs.ErrInvalidContactWebhook
	}
	return nil
}

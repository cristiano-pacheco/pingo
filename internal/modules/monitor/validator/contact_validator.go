package validator

import (
	"net/mail"
	"net/url"

	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/enum"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/errs"
)

type ContactValidator interface {
	Validate(contactType, contactData string) error
}

type contactValidator struct {
}

func NewContactValidator() ContactValidator {
	return &contactValidator{}
}

func (v *contactValidator) Validate(contactType, contactData string) error {
	switch contactType {
	case enum.ContactTypeEmail:
		return v.validateEmail(contactData)
	case enum.ContactTypeWebhook:
		return v.validateWebhook(contactData)
	}
	return nil
}

func (v *contactValidator) validateEmail(contactData string) error {
	_, err := mail.ParseAddress(contactData)
	if err != nil {
		return errs.ErrInvalidContactEmail
	}
	return nil
}

func (v *contactValidator) validateWebhook(contactData string) error {
	parsedURL, err := url.ParseRequestURI(contactData)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return errs.ErrInvalidContactWebhook
	}
	return nil
}

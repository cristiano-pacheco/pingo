package service

import (
	"bytes"
	"html/template"
)

type EmailTemplateServiceI interface {
	CompileAccountConfirmationTemplate(input AccountConfirmationInput) (string, error)
	CompileAuthVerificationCodeTemplate(name string, code string) (string, error)
}

type EmailTemplateService struct {
}

var _ EmailTemplateServiceI = (*EmailTemplateService)(nil)

func NewEmailTemplateService() *EmailTemplateService {
	return &EmailTemplateService{}
}

type AccountConfirmationInput struct {
	Name                    string
	AccountConfirmationLink string
}

func (s *EmailTemplateService) CompileAccountConfirmationTemplate(input AccountConfirmationInput) (string, error) {
	// Load templates
	tmpl, err := template.New("layout_default.gohtml").
		ParseFiles(
			"internal/modules/identity/ui/email/templates/layout_default.gohtml",
			"internal/modules/identity/ui/email/templates/account_confirmation.gohtml",
		)
	if err != nil {
		return "", err
	}

	// Prepare data
	data := map[string]interface{}{
		"Name":                    input.Name,
		"AccountConfirmationLink": input.AccountConfirmationLink,
		"Title":                   "Account Confirmation",
	}

	// Render template
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "htmlBody", data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *EmailTemplateService) CompileAuthVerificationCodeTemplate(name string, code string) (string, error) {
	// Load templates
	tmpl, err := template.New("layout_default.gohtml").
		ParseFiles(
			"internal/modules/identity/ui/email/templates/layout_default.gohtml",
			"internal/modules/identity/ui/email/templates/auth_verification_code.gohtml",
		)
	if err != nil {
		return "", err
	}

	// Prepare data
	data := map[string]interface{}{
		"Name":  name,
		"Code":  code,
		"Title": "Auth Verification Code",
	}

	// Render template
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "htmlBody", data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

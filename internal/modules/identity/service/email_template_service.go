package service

import (
	"bytes"
	"context"
	"html/template"
)

type EmailTemplateService interface {
	CompileAccountConfirmationTemplate(ctx context.Context, input AccountConfirmationInput) (string, error)
	CompileAuthVerificationCodeTemplate(ctx context.Context, name string, code string) (string, error)
}

type emailTemplateService struct {
}

func NewEmailTemplateService() EmailTemplateService {
	return &emailTemplateService{}
}

type AccountConfirmationInput struct {
	Name                    string
	AccountConfirmationLink string
}

func (s *emailTemplateService) CompileAccountConfirmationTemplate(
	ctx context.Context,
	input AccountConfirmationInput,
) (string, error) {
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

func (s *emailTemplateService) CompileAuthVerificationCodeTemplate(
	ctx context.Context,
	name string,
	code string,
) (string, error) {
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

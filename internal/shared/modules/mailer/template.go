package mailer

import (
	"bytes"
	"embed"
	"html/template"
)

type Template interface {
	CompileTemplate(input CompileTemplateInput) (string, error)
	CompileBlankTemplate(input CompileTemplateInput) (string, error)
}

type CompileTemplateInput struct {
	TemplateName        string
	LayoutTpl           string
	TemplatePath        string
	TemplateSectionName string
	TemplateFS          embed.FS
	Data                any
}

type mailerTemplate struct {
}

func NewMailerTemplate() Template {
	return &mailerTemplate{}
}

func (mt *mailerTemplate) CompileTemplate(
	input CompileTemplateInput,
) (string, error) {
	tmpl, err := template.New(input.TemplateName).ParseFS(input.TemplateFS, input.LayoutTpl, input.TemplatePath)
	if err != nil {
		return "", err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, input.TemplateSectionName, input.Data)
	if err != nil {
		return "", err
	}

	return htmlBody.String(), nil
}

func (mt *mailerTemplate) CompileBlankTemplate(
	input CompileTemplateInput,
) (string, error) {
	tmpl, err := template.New(input.TemplateName).ParseFS(input.TemplateFS, input.TemplatePath)
	if err != nil {
		return "", err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, input.TemplateSectionName, input.Data)
	if err != nil {
		return "", err
	}

	return htmlBody.String(), nil
}

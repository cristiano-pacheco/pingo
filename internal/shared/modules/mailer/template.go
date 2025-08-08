package mailer

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed "templates"
var templateFS embed.FS

type Template interface {
	CompileTemplate(templateName string, data any) (string, error)
	CompileBlankTemplate(templateName string, data any) (string, error)
}

type mailerTemplate struct {
}

func NewMailerTemplate() Template {
	return &mailerTemplate{}
}

func (mt *mailerTemplate) CompileTemplate(templateName string, data any) (string, error) {
	layoutTpl := "templates/layout/default.gohtml"
	tmpl, err := template.New("email").ParseFS(templateFS, layoutTpl, "templates/"+templateName)
	if err != nil {
		return "", err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return "", err
	}

	return htmlBody.String(), nil
}

func (mt *mailerTemplate) CompileBlankTemplate(templateName string, data any) (string, error) {
	tmpl, err := template.New("email").ParseFS(templateFS, templateName)
	if err != nil {
		return "", err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return "", err
	}

	return htmlBody.String(), nil
}

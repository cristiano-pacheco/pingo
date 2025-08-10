package templates

import "embed"

//go:embed *.gohtml
var EmailTemplates embed.FS

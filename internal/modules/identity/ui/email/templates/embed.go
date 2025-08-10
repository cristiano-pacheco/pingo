package templates

import "embed"

//go:embed *.gohtml
var EmailTemplatesFS embed.FS

package web

import (
    "embed"
    "html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var Templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))

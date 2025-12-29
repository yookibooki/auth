package web

import (
	"html/template"
	"log"
)

func Parse() *template.Template {
	tmpl, err := template.ParseGlob("web/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	return tmpl
}

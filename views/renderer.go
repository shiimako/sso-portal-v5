// file: views/renderer.go
package views

import (
	"html/template"
	"strings"
)

// InitTemplates menginisialisasi dan mem-parsing semua file template HTML.
func InitTemplates() (*template.Template, error) {
	// Definisikan custom functions
	funcMap := template.FuncMap{
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
	}

	// Buat template baru dengan funcMap
	tmpl := template.New("base").Funcs(funcMap)

	// Parse layouts
	tmpl, err := tmpl.ParseGlob("views/layouts/*.html")
	if err != nil {
		return nil, err
	}

	// Parse pages
	patterns := []string{
		"views/pages/*.html",
		"views/pages/*/*.html",
		"views/pages/*/*/*.html",
	}
	for _, pattern := range patterns {
		tmpl.ParseGlob(pattern)
	}

	// (Opsional) partials/components
	// _, err = tmpl.ParseGlob("views/components/*.html")
	// if err != nil {
	// 	return nil, err
	// }

	return tmpl, nil
}

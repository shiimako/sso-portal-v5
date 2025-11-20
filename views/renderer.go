// file: views/renderer.go
package views

import (
	"html/template"
	"log"
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
		_, err := tmpl.ParseGlob(pattern)
		if err != nil {
			// Jika ada 1 file saja yang gagal di-parse,
			// hentikan aplikasi dan beri tahu file mana yang error.
			log.Printf("ERROR: Gagal parsing template pattern '%s': %v", pattern, err)
			return nil, err
		}
	}

	// (Opsional) partials/components
	// _, err = tmpl.ParseGlob("views/components/*.html")
	// if err != nil {
	// 	return nil, err
	// }

	return tmpl, nil
}

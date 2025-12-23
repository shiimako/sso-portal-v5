// file: views/renderer.go

package views

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sso-portal-v5/config"
	"sso-portal-v5/models"
	"strings"
)

type ViewData struct {
	UserLogin   *models.FullUser
	ActiveRole  string
	HeaderTitle string
	Data        map[string]interface{}
}

type Views struct {
	env *config.Env
}

func NewViews(env *config.Env) *Views {
	return &Views{env: env}
}

// InitTemplates mengembalikan map, dimana key adalah nama halaman (misal: "dashboard")
func InitTemplates() (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	funcMap := template.FuncMap{
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"js": func(s string) template.JS { return template.JS(s) },
		"json": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}


	rootTmpl := template.New("root").Funcs(funcMap)

	patterns := []string{
		"views/templates/layouts/*.html",
		"views/templates/partials/*.html",
	}

	for _, pattern := range patterns {
		if _, err := rootTmpl.ParseGlob(pattern); err != nil {
			return nil, err
		}
	}

	// Muat semua halaman dari views/pages/
	pagePatterns := []string{
		"views/pages/*.html",     // Untuk file langsung di pages/
		"views/pages/*/*.html",   // Untuk pages/{folder}/{page}.html
		"views/pages/*/*/*.html", // Untuk pages/{folder}/{subfolder}/{page}.html
	}

	var pages []string
	for _, pattern := range pagePatterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		pages = append(pages, matches...)
	}

	if len(pages) == 0 {
		log.Println("WARNING: Tidak ada file template halaman yang ditemukan!")
	}


	for _, pageFile := range pages {
		fileName := filepath.Base(pageFile)
		name := strings.TrimSuffix(fileName, ".html")

		tmplClone, err := rootTmpl.Clone()
		if err != nil {
			return nil, err
		}

		_, err = tmplClone.ParseFiles(pageFile)
		if err != nil {
			return nil, err
		}

		// Simpan template dengan nama halaman sebagai key
		templates[name] = tmplClone
	}
	return templates, nil
}

func (v *Views) RenderPage(w http.ResponseWriter, r *http.Request, name string, pageData map[string]interface{}) {

	userLogin := r.Context().Value("UserLogin").(*models.FullUser)
	activeRole := userLogin.Roles[0].Name

	data := ViewData{
		UserLogin:   userLogin,
		ActiveRole:  activeRole,
		HeaderTitle: "PNC-Portal System",
		Data:        pageData,
	}

	tmpl := v.env.Templates[name]
	tmpl.ExecuteTemplate(w, "base", data)
}

// file: views/renderer.go

package views

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sso-portal-v3/config"
	"sso-portal-v3/models"
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
		"js": func(s string) template.JS { return template.JS(s) }, // Helper untuk JS di dashboard
	}

	// 1. Parse Layouts & Partials dulu (Base, Header, dll) ke dalam template 'root'
	// Pastikan base.html ada di folder layouts
	rootTmpl := template.New("root").Funcs(funcMap)

	// Parse glob layouts & partials
	patterns := []string{
		"views/templates/layouts/*.html",
		"views/templates/partials/*.html",
	}

	for _, pattern := range patterns {
		if _, err := rootTmpl.ParseGlob(pattern); err != nil {
			return nil, err
		}
	}

	// 2. Loop setiap file Halaman secara manual
	// Ini kuncinya: Setiap halaman punya instance template sendiri
	pagePatterns := []string{
		"views/pages/*.html",     // Untuk file langsung di pages/ (jika ada)
		"views/pages/*/*.html",   // Untuk pages/auth/login.html, pages/dashboard/dashboard.html
		"views/pages/*/*/*.html", // Untuk pages/admin/apps/admin-app-list.html
	}

	var pages []string
	for _, pattern := range pagePatterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		pages = append(pages, matches...)
	}

	// Cek jika tidak ada halaman ditemukan sama sekali (opsional, buat debugging)
	if len(pages) == 0 {
		log.Println("WARNING: Tidak ada file template halaman yang ditemukan!")
	}

	// Jika ada subfolder (views/pages/*/*.html), tambahkan logic glob tambahan disini

	for _, pageFile := range pages {
		fileName := filepath.Base(pageFile)
		name := strings.TrimSuffix(fileName, ".html") // misal: "dashboard"

		// Clone dari root (mendapatkan base, header, dll)
		tmplClone, err := rootTmpl.Clone()
		if err != nil {
			return nil, err
		}

		// Parse file halaman spesifik ke dalam clone tersebut
		_, err = tmplClone.ParseFiles(pageFile)
		if err != nil {
			return nil, err
		}

		// Simpan ke map dengan nama file sebagai key (misal: "dashboard")
		// Catatan: Di controller nanti panggil berdasarkan key ini, bukan define name.
		templates[name] = tmplClone
		// Atau jika file mendefinisikan {{define "dashboard"}}, Anda tetap bisa pakai key string bebas.
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

// file: controllers/admincontroller/admincontroller_app.go

package admincontroller

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sso-portal-v5/models"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ListApplications menampilkan halaman daftar semua aplikasi.
func (ac *AdminController) ListApplications(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	flashes := session.Flashes()
	session.Save(r, w)

	var flashMsg string
	if len(flashes) > 0 {
		flashMsg = flashes[0].(string)
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")

	page := 1
	limit := 10

	if pageStr != "" {
		p, _ := strconv.Atoi(pageStr)
		if p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		l, _ := strconv.Atoi(limitStr)
		if l > 0 {
			limit = l
		}
	}

	apps, err := models.GetAllApplications(ac.env.DB, page, limit, search)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	pageData := map[string]interface{}{
		"Apps":   apps,
		"Page":   page,
		"Limit":  limit,
		"Search": search,
		"Flash":  flashMsg,
	}

	ac.views.RenderPage(w, r, "admin-app-list", pageData)
}

// DetailApplication menampilkan halaman detail read-only untuk sebuah aplikasi.
func (ac *AdminController) DetailApplication(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	app, roleIDs, PosIDs, err := models.FindApplicationByID(ac.env.DB, id)
	if err != nil {
		ac.RenderError(w, r, http.StatusNotFound, "Aplikasi tidak ditemukan")
		return
	}

	var roleNames []string
	var posNames []string
	if len(roleIDs) > 0 {
		query, args, _ := sqlx.In("SELECT role_name FROM roles WHERE id IN (?)", roleIDs)
		query = ac.env.DB.Rebind(query)
		err = ac.env.DB.Select(&roleNames, query, args...)
		if err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
			return
		}
	}

	if len(PosIDs) > 0 {
		query, args, _ := sqlx.In("SELECT position_name FROM positions WHERE id IN (?)", PosIDs)
		query = ac.env.DB.Rebind(query)
		err = ac.env.DB.Select(&posNames, query, args...)
		if err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
			return
		}
	}

	data := map[string]interface{}{
		"App":            app,
		"RoleNames":      roleNames,
		"PositionsNames": posNames,
	}

	ac.views.RenderPage(w, r, "admin-app-detail", data)
}

// NewApplicationForm menampilkan form untuk membuat aplikasi baru.
func (ac *AdminController) NewApplicationForm(w http.ResponseWriter, r *http.Request) {
	roles, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	position, err := models.GetAllPositions(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	categories, err := models.GetAllCategories(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	data := map[string]interface{}{
		"Roles":    roles,
		"Position": position,
		"Categories": categories,
	}

	ac.views.RenderPage(w, r, "admin-app-form", data)
}

// CreateApplication memproses form pembuatan aplikasi baru.
func (ac *AdminController) CreateApplication(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File terlalu besar atau form error", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	slug := r.FormValue("slug")
	targetURL := r.FormValue("target_url")
	iconURL := "/uploads/icons/default.png"
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))

	roleIDs := r.Form["role_ids"]
	posIDs := r.Form["position_ids"]

	if name == "" || slug == "" || targetURL == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama, Slug, dan Target URL harus diisi")
		return
	}

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		ac.RenderError(w, r, http.StatusBadRequest, "URL Tujuan harus diawali http:// atau https://")
		return
	}

	file, header, err := r.FormFile("icon-file")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{
			".png": true, ".jpg": true, ".jpeg": true, ".svg": true,
		}

		if !allowed[ext] {
			ac.RenderError(w, r, http.StatusBadRequest, "Format File harus .png, atau .jpeg, atau .svg")
			return
		}

		// Buat filename : app_{slug}_icon.png
		safeSlug := strings.ReplaceAll(slug, " ", "-")
		filename := fmt.Sprintf("app-%s-icon%s", safeSlug, ext)
		savePath := filepath.Join("public", "uploads", "icons", filename)

		// Buat folder jika belum ada
		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
			return
		}

		out, err := os.Create(savePath)
		if err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
			return
		}

		iconURL = "/uploads/icons/" + filename
	}

	err = models.CreateApplication(ac.env.DB, name, description, slug, targetURL, iconURL, categoryID, roleIDs, posIDs)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Aplikasi " + name + " berhasil ditambahkan!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
}

// EditApplicationForm menampilkan form untuk mengedit aplikasi.
func (ac *AdminController) EditApplicationForm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Ambil data aplikasi yang akan diedit
	app, currentRoleIDs, currentPosIDs, err := models.FindApplicationByID(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Aplikasi Tidak Ditemukan")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	allRoles, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	allPos, err := models.GetAllPositions(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	categories, err := models.GetAllCategories(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	currentRolesMap := make(map[int]bool)
	for _, rid := range currentRoleIDs {
		currentRolesMap[rid] = true
	}

	currentPosMap := make(map[int]bool)
	for _, rid := range currentPosIDs {
		currentPosMap[rid] = true
	}

	data := map[string]interface{}{
		"App":              app,
		"AllRoles":         allRoles,
		"AllPositions":     allPos,
		"CurrentRoles":     currentRolesMap,
		"CurrentPositions": currentPosMap,
		"Categories":       categories,
	}

	ac.views.RenderPage(w, r, "admin-app-edit", data)
}

// UpdateApplication memproses form edit aplikasi.
func (ac *AdminController) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "File Terlalu Besar, maksimal 10 MB")
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	slug := r.FormValue("slug")
	targetURL := r.FormValue("target_url")
	iconURL := r.FormValue("icon_url")
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))

	roleIDs := r.Form["role_ids"]
	posIDs := r.Form["position_ids"]

	if name == "" || slug == "" || targetURL == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama, Slug. dan Target URL Harus Diisi.")
		return
	}

	file, header, err := r.FormFile("icon-file")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".svg": true}
		if !allowed[ext] {
			ac.RenderError(w, r, http.StatusBadRequest, "Format File harus .png, atau .jpeg, atau .svg")
			return
		}

		safeSlug := strings.ReplaceAll(slug, " ", "-")
		filename := fmt.Sprintf("app-%s-icon%s", safeSlug, ext)
		savePath := filepath.Join("public", "uploads", "icons", filename)

		os.MkdirAll(filepath.Dir(savePath), 0755)

		out, err := os.Create(savePath)
		if err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
			return
		}

		iconURL = "/uploads/icons/" + filename
	}

	err = models.UpdateApplication(ac.env.DB, id, name, description, slug, targetURL, iconURL, categoryID, roleIDs, posIDs)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Aplikasi " + name + " berhasil diperbarui!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
}

// DeleteApplication menghapus aplikasi berdasarkan ID.
func (ac *AdminController) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	err := models.DeleteApplication(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Aplikasi Tidak Ditemukan.")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Aplikasi dengan id " + id + " berhasil dihapus!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
}

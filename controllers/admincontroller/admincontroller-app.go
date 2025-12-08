// file: controllers/admincontroller/admincontroller_app.go

package admincontroller

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sso-portal-v3/models"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ListApplications menampilkan halaman daftar semua aplikasi.
func (ac *AdminController) ListApplications(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, "Gagal mengambil data aplikasi", http.StatusInternalServerError)
		return
	}

	pageData := map[string]interface{}{
		"Apps":   apps,
		"Page":   page,
		"Limit":  limit,
		"Search": search,
	}

	ac.views.RenderPage(w, r, "admin-app-list", pageData)
}

// DetailApplication menampilkan halaman detail read-only untuk sebuah aplikasi.
func (ac *AdminController) DetailApplication(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	app, roleIDs, PosIDs, err := models.FindApplicationByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "Data aplikasi tidak ditemukan", http.StatusNotFound)
		return
	}

	var roleNames []string
	var posNames []string
	if len(roleIDs) > 0 {
		query, args, _ := sqlx.In("SELECT role_name FROM roles WHERE id IN (?)", roleIDs)
		query = ac.env.DB.Rebind(query)
		err = ac.env.DB.Select(&roleNames, query, args...)
		if err != nil {
			log.Printf("ERROR: Gagal mengambil nama peran: %v", err)
		}
	}

	if len(PosIDs) > 0 {
		query, args, _ := sqlx.In("SELECT position_name FROM positions WHERE id IN (?)", PosIDs)
		query = ac.env.DB.Rebind(query)
		err = ac.env.DB.Select(&posNames, query, args...)
		if err != nil {
			log.Printf("ERROR: Gagal mengambil nama posisi: %v", err)
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
		http.Error(w, "Gagal mengambil data peran", http.StatusInternalServerError)
		return
	}

	position, err := models.GetAllPositions(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data peran", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Roles":    roles,
		"Position": position,
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

	roleIDs := r.Form["role_ids"]
	posIDs := r.Form["position_ids"]

	if name == "" || slug == "" || targetURL == "" {
		http.Error(w, "Nama, Slug, dan Target URL wajib diisi", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		http.Error(w, "Target URL harus diawali dengan http:// atau https://", http.StatusBadRequest)
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
			http.Error(w, "Format file icon tidak didukung", http.StatusBadRequest)
			return
		}

		// Buat filename aman: app_{slug}_icon.png
		safeSlug := strings.ReplaceAll(slug, " ", "-")
		filename := fmt.Sprintf("app-%s-icon%s", safeSlug, ext)

		// Path penyimpanan
		savePath := filepath.Join("public", "uploads", "icons", filename)

		// Buat folder jika belum ada
		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			log.Println("ERROR: Gagal buat folder icons:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		out, err := os.Create(savePath)
		if err != nil {
			log.Println("ERROR: Gagal membuat file icon:", err)
			http.Error(w, "Gagal menyimpan icon aplikasi", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			log.Println("ERROR: Gagal menulis file icon:", err)
			http.Error(w, "Gagal menyimpan icon aplikasi", http.StatusInternalServerError)
			return
		}

		iconURL = "/uploads/icons/" + filename
	}

	err = models.CreateApplication(ac.env.DB, name, description, slug, targetURL, iconURL, roleIDs, posIDs)
	if err != nil {
		log.Printf("ERROR: Gagal menyimpan aplikasi baru: %v", err)
		http.Error(w, "Gagal menyimpan data aplikasi ke database", http.StatusInternalServerError)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Aplikasi baru berhasil ditambahkan!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
}

// EditApplicationForm menampilkan form untuk mengedit aplikasi.
func (ac *AdminController) EditApplicationForm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Ambil data aplikasi yang akan diedit
	app, currentRoleIDs, currentPosIDs, err := models.FindApplicationByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "Data aplikasi tidak ditemukan", http.StatusNotFound)
		return
	}

	allRoles, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		log.Println("Error : ", err)
		http.Error(w, "Gagal mengambil data peran", http.StatusInternalServerError)
		return
	}

	allPos, err := models.GetAllPositions(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data posisi", http.StatusInternalServerError)
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
	}

	ac.views.RenderPage(w, r, "admin-app-edit", data)
}

// UpdateApplication memproses form edit aplikasi.
func (ac *AdminController) UpdateApplication(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]

    if err := r.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "Form error atau file terlalu besar", http.StatusBadRequest)
        return
    }

    name := r.FormValue("name")
    description := r.FormValue("description")
    slug := r.FormValue("slug")
    targetURL := r.FormValue("target_url")
    iconURL := r.FormValue("icon_url") 
    
    roleIDs := r.Form["role_ids"]
    posIDs := r.Form["position_ids"]

    if name == "" || slug == "" || targetURL == "" {
        http.Error(w, "Semua field wajib diisi", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("icon-file")
    if err == nil {
        defer file.Close()


        ext := strings.ToLower(filepath.Ext(header.Filename))
        allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".svg": true}
        if !allowed[ext] {
            http.Error(w, "Format file tidak didukung", http.StatusBadRequest)
            return
        }


        safeSlug := strings.ReplaceAll(slug, " ", "-")
        filename := fmt.Sprintf("app-%s-icon%s", safeSlug, ext)
        savePath := filepath.Join("public", "uploads", "icons", filename)


        os.MkdirAll(filepath.Dir(savePath), 0755)

        out, err := os.Create(savePath)
        if err != nil {
            log.Println("ERROR Create File:", err)
            http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
            return
        }
        defer out.Close()

        if _, err := io.Copy(out, file); err != nil {
            log.Println("ERROR Copy File:", err)
            http.Error(w, "Gagal menulis file", http.StatusInternalServerError)
            return
        }

        iconURL = "/uploads/icons/" + filename
    }

    err = models.UpdateApplication(ac.env.DB, id, name, description, slug, targetURL, iconURL, roleIDs, posIDs)
    if err != nil {
        log.Printf("ERROR: Gagal update aplikasi: %v", err)
        http.Error(w, "Gagal memperbarui data aplikasi", http.StatusInternalServerError)
        return
    }

    session, _ := ac.env.Store.Get(r, ac.env.SessionName)
    session.AddFlash("Aplikasi berhasil diperbarui!")
    session.Save(r, w)

    http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
}

// DeleteApplication menghapus aplikasi berdasarkan ID.
func (ac *AdminController) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	err := models.DeleteApplication(ac.env.DB, id)
	if err != nil {
		log.Printf("ERROR: Gagal menghapus aplikasi: %v", err)
		http.Error(w, "Gagal menghapus aplikasi", http.StatusInternalServerError)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Aplikasi berhasil dihapus!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
}

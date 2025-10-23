// file: controllers/admincontroller/admincontroller_app.go

package admincontroller

import (
	"log"
	"net/http"
	"sso-portal-v3/models"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ListApplications menampilkan halaman daftar semua aplikasi.
func (ac *AdminController) ListApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := models.GetAllApplications(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data aplikasi", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Applications": apps,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-app-list.html", data)
}

// DetailApplication menampilkan halaman detail read-only untuk sebuah aplikasi.
func (ac *AdminController) DetailApplication(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Gunakan model yang sudah ada untuk mengambil detail aplikasi dan ID perannya
	app, roleIDs, err := models.FindApplicationByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "Data aplikasi tidak ditemukan", http.StatusNotFound)
		return
	}

	var roleNames []string
	if len(roleIDs) > 0 {
		// Ambil nama peran berdasarkan ID yang didapat
		query, args, _ := sqlx.In("SELECT name FROM roles WHERE id IN (?)", roleIDs)
		query = ac.env.DB.Rebind(query)
		err = ac.env.DB.Select(&roleNames, query, args...)
		if err != nil {
			log.Printf("ERROR: Gagal mengambil nama peran: %v", err)
			// Jangan hentikan proses, tampilkan saja aplikasinya tanpa peran
		}
	}


	data := map[string]interface{}{
		"App":       app,
		"RoleNames": roleNames, // Kirim daftar nama peran
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-app-detail.html", data)
}

// NewApplicationForm menampilkan form untuk membuat aplikasi baru.
func (ac *AdminController) NewApplicationForm(w http.ResponseWriter, r *http.Request) {
	// Panggil model untuk mendapatkan semua peran
	roles, err := models.GetAllRolesForAccess(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data peran", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Roles": roles,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-app-form.html", data)
}

// CreateApplication memproses form pembuatan aplikasi baru.
func (ac *AdminController) CreateApplication(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	slug := r.FormValue("slug")
	targetURL := r.FormValue("target_url")
	roleIDs := r.Form["role_ids"] // Mengambil semua nilai checkbox sebagai slice

	// Validasi sederhana
	if name == "" || slug == "" || targetURL == "" {
		http.Error(w, "Nama, Slug, dan Target URL wajib diisi", http.StatusBadRequest)
		return
	}

	// Panggil model untuk menyimpan data
	err := models.CreateApplication(ac.env.DB, name, slug, targetURL, roleIDs)
	if err != nil {
		log.Printf("ERROR: Gagal menyimpan aplikasi baru: %v", err)
		http.Error(w, "Gagal menyimpan data aplikasi ke database", http.StatusInternalServerError)
		return
	}

	// Beri pesan sukses dan redirect
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Aplikasi baru berhasil ditambahkan!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/applications", http.StatusSeeOther)
}

// EditApplicationForm menampilkan form untuk mengedit aplikasi.
func (ac *AdminController) EditApplicationForm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// Ambil data aplikasi yang akan diedit
	app, currentRoleIDs, err := models.FindApplicationByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "Data aplikasi tidak ditemukan", http.StatusNotFound)
		return
	}

	// Ambil semua kemungkinan peran untuk ditampilkan sebagai checkbox
	allRoles, err := models.GetAllRolesForAccess(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data peran", http.StatusInternalServerError)
		return
	}
	
	// Buat map untuk memudahkan pengecekan di template (mana yang harus dicentang)
	currentRolesMap := make(map[int]bool)
	for _, rid := range currentRoleIDs {
		currentRolesMap[rid] = true
	}

	data := map[string]interface{}{
		"App":          app,
		"AllRoles":     allRoles,
		"CurrentRoles": currentRolesMap,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-app-edit.html", data)
}

// UpdateApplication memproses form edit aplikasi.
func (ac *AdminController) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	slug := r.FormValue("slug")
	targetURL := r.FormValue("target_url")
	roleIDs := r.Form["role_ids"] // Mengambil semua nilai checkbox

	if name == "" || slug == "" || targetURL == "" {
		http.Error(w, "Semua field wajib diisi", http.StatusBadRequest)
		return
	}

	err := models.UpdateApplication(ac.env.DB, id, name, slug, targetURL, roleIDs)
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
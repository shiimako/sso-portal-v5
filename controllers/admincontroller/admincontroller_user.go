package admincontroller

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sso-portal-v3/models"
	"strings"

	"github.com/gorilla/mux"
)

// ListUsers menampilkan daftar semua pengguna.
func (ac *AdminController) ListUsers(w http.ResponseWriter, r *http.Request) {

	users, err := models.GetAllUsers(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data pengguna", http.StatusInternalServerError)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	flashes := session.Flashes()
	session.Save(r, w)

	data := map[string]interface{}{
		"Users":   users,
		"Flashes": flashes,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-user-list.html", data)
}

// DetailUser menampilkan halaman detail read-only untuk seorang pengguna.
func (ac *AdminController) DetailUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "ID pengguna tidak ditemukan", http.StatusBadRequest)
		return
	}

	// Gunakan kembali fungsi yang sama dengan form edit
	user, err := models.FindUserByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal mengambil data pengguna: %v", err), http.StatusInternalServerError)
		return
	}

	// Kita juga perlu mengambil daftar nama peran untuk ditampilkan
	var roles []string
	queryRoles := `SELECT r.name FROM roles r 
	               JOIN user_roles ur ON r.id = ur.role_id 
	               WHERE ur.user_id = ?`
	rows, err := ac.env.DB.Query(queryRoles, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var roleName string
			if err := rows.Scan(&roleName); err == nil {
				roles = append(roles, roleName)
			}
		}
	}

	data := map[string]interface{}{
		"User":  user,
		"Roles": roles, // Kirim daftar nama peran yang dimiliki
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-user-detail.html", data)
}

// NewUserForm menampilkan form untuk membuat pengguna baru.
func (ac *AdminController) NewUserForm(w http.ResponseWriter, r *http.Request) {

	// Ambil semua peran dasar (selain admin) dari database untuk ditampilkan di dropdown
	roles, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data peran", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Roles": roles,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-user-form.html", data)
}

// CreateUser memproses form pembuatan pengguna baru.
func (ac *AdminController) CreateUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal memproses form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	roleID := r.FormValue("role_id")
	nim := r.FormValue("nim")
	nip := r.FormValue("nip")
	nidn := r.FormValue("nidn")
	address := r.FormValue("address")
	phone := r.FormValue("phone_number")

	if name == "" || email == "" || roleID == "" || address == "" || phone == "" {
		http.Error(w, "Semua field wajib diisi", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(email, "@pnc.ac.id") {
		http.Error(w, "Email harus menggunakan domain @pnc.ac.id", http.StatusBadRequest)
		return
	}

	switch roleID {
	case "2": // dosen
		if nip == "" || nidn == "" {
			http.Error(w, "NIP dan NIDN wajib diisi untuk dosen", http.StatusBadRequest)
			return
		}
	case "3": // mahasiswa
		if nim == "" {
			http.Error(w, "NIM wajib diisi untuk mahasiswa", http.StatusBadRequest)
			return
		}
	}

	err := models.StoreNewUser(ac.env.DB, name, email, roleID, nip, nidn, nim, address, phone)
	if err != nil {
		http.Error(w, "Gagal menyimpan pengguna baru", http.StatusInternalServerError)
		log.Println("ERROR: Gagal menyimpan pengguna baru:", err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Pengguna berhasil ditambahkan!!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// EditUserForm menampilkan form untuk mengedit pengguna yang sudah ada.
func (ac *AdminController) EditUserForm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "ID pengguna tidak ditemukan", http.StatusBadRequest)
		return
	}

	// 1. Panggil model untuk mendapatkan data user spesifik
	user, err := models.FindUserByID(ac.env.DB, id) // Menggunakan fungsi baru yang lebih efisien
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal mengambil data pengguna: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Panggil model (lagi) untuk mendapatkan semua peran yang bisa dipilih (untuk checkbox)
	allRoles, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal mengambil daftar peran: %v", err), http.StatusInternalServerError)
		return
	}

	//3. Panggil model untuk mendapatkan semua atribut yang bisa dipilih (untuk checkbox)
	allAttributes, err := models.GetAllAttributes(ac.env.DB)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal mengambil daftar atribut: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. Gabungkan keduanya untuk dikirim ke template
	data := map[string]interface{}{
		"User":          user,
		"BaseRoles":     allRoles,
		"AllAttributes": allAttributes,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-user-edit.html", data)
}

// UpdateUser memproses form edit pengguna
func (ac *AdminController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "ID pengguna tidak ditemukan", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	status := r.FormValue("status")

	baseroleIDs := r.Form["role_ids"]
	attributeIDs := r.Form["attribute_ids"]

	attributeScopes := make(map[string]string)
	for key, values := range r.Form {
		if strings.HasPrefix(key, "scope_") {
			// Dapatkan ID atribut dari nama field, misal "scope_4" -> "4"
			attributeID := strings.TrimPrefix(key, "scope_")
			// Simpan scope-nya, hanya jika checkbox-nya juga dicentang
			if contains(attributeIDs, attributeID) {
				attributeScopes[attributeID] = values[0]
			}
		}
	}

	nim := r.FormValue("nim")
	nip := r.FormValue("nip")
	nidn := r.FormValue("nidn")

	if name == "" || email == "" || status == "" || len(baseroleIDs) == 0 {
		http.Error(w, "Semua field wajib diisi", http.StatusBadRequest)
		return
	}

	// Ambil email admin dari .env
	adminEmail := os.Getenv("ADMIN_EMAIL_OVERRIDE")

	// Cek apakah email yang dimasukkan adalah email @pnc.ac.id
	isPncEmail := strings.HasSuffix(email, "@pnc.ac.id")

	// Cek apakah email yang dimasukkan adalah email admin override
	isAdminOverride := (adminEmail != "" && email == adminEmail)

	if !isPncEmail && !isAdminOverride {
		http.Error(w, "Email harus menggunakan domain @pnc.ac.id", http.StatusBadRequest)
		return
	}

	if contains(baseroleIDs, "2") { // dosen
		if nip == "" || nidn == "" {
			http.Error(w, "NIP dan NIDN wajib diisi untuk dosen", http.StatusBadRequest)
			return
		}
	}
	if contains(baseroleIDs, "3") { // mahasiswa
		if nim == "" {
			http.Error(w, "NIM wajib diisi untuk mahasiswa", http.StatusBadRequest)
			return
		}
	}

	// Panggil model dengan parameter yang sudah lengkap
	err := models.UpdateUser(ac.env.DB, id, name, email, status, baseroleIDs, attributeIDs, attributeScopes, nim, nip, nidn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal memperbarui data pengguna: %v", err), http.StatusInternalServerError)
		log.Println("ERROR: Gagal memperbarui data pengguna:", err)
		return
	}

	sessions, _ := ac.env.Store.Get(r, ac.env.SessionName)
	sessions.AddFlash("Data pengguna berhasil diperbarui!!")
	sessions.Save(r, w)

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// DeleteUser menghapus user berdasarkan ID
func (ac *AdminController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "ID pengguna tidak ditemukan", http.StatusBadRequest)
		return
	}
	err := models.DeleteUser(ac.env.DB, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal menghapus pengguna: %v", err), http.StatusInternalServerError)
		log.Println("ERROR: Gagal menghapus pengguna:", err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Pengguna berhasil dihapus!!")
	session.Save(r, w)
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// Helper cek role_id ada di slice
func contains(list []string, val string) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

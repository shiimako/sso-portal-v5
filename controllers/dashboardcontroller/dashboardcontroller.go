// file: controllers/dashboardcontroller/dashboardcontroller.go

package dashboardcontroller

import (
	"fmt"
	"log"
	"net/http"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"

	"github.com/jmoiron/sqlx"
)

// Definisikan struct untuk data aplikasi
type Application struct {
	Name      string
	Slug      string
	TargetURL string
}

type DashboardController struct {
	env *handlers.Env
}

func NewDashboardController(env *handlers.Env) *DashboardController {
	return &DashboardController{env: env}
}

func (dc *DashboardController) Index(w http.ResponseWriter, r *http.Request) {
	session, _ := dc.env.Store.Get(r, dc.env.SessionName)

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		// Jika user_id tidak ada, berarti belum login / session expired
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user, err := models.FindUserByID(dc.env.DB, fmt.Sprintf("%d", userID)) // Gunakan user_id
	if err != nil {
		log.Printf("ERROR: Gagal mengambil user %d dari DB: %v", userID, err)
		// Hancurkan session jika user tidak ditemukan di DB
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Error(w, "Gagal memuat data pengguna.", http.StatusInternalServerError)
		return
	}

	if user.Status != "aktif" {
		log.Printf("INFO: Akses dashboard ditolak untuk user %d karena status: %s", userID, user.Status)
		// Hancurkan session jika user tidak aktif
		session.Options.MaxAge = -1
		session.Save(r, w)
		// Tampilkan pesan error yang sesuai
		http.Error(w, fmt.Sprintf("Akun Anda berstatus '%s'. Tidak dapat mengakses dashboard.", user.Status), http.StatusForbidden)
		return
	}

	allRolesFromDB, err := models.GetUserRolesAndAttributes(dc.env.DB, userID)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil peran terbaru untuk user %d: %v", userID, err)
		http.Error(w, "Gagal mengambil data peran.", http.StatusInternalServerError)
		return
	}

	var currentBaseRoles []string
	var currentAttributes []map[string]interface{}
	var allCurrentRoleNames []string // Untuk query aplikasi

	for _, role := range allRolesFromDB {
		allCurrentRoleNames = append(allCurrentRoleNames, role.Name)
		if role.Type == "base" {
			currentBaseRoles = append(currentBaseRoles, role.Name)
		} else if role.Type == "attribute" {
			attr := map[string]interface{}{"role": role.Name}
			if role.Scope.Valid {
				attr["scope"] = role.Scope.String
			}
			currentAttributes = append(currentAttributes, attr)
		}
	}

	// Cek apakah peran aktif sudah dipilih
	activeRoleFromSession, sessionHasActiveRole := session.Values["active_role"].(string)
	// Validasi: Apakah peran aktif dari session masih dimiliki oleh user?
	isValidActiveRole := false
	if sessionHasActiveRole && activeRoleFromSession != "" {
		for _, dbRole := range currentBaseRoles {
			if dbRole == activeRoleFromSession {
				isValidActiveRole = true
				break
			}
		}
	}

	var finalActiveRole string // Peran aktif yang akan digunakan

	if !isValidActiveRole {
		// Jika peran aktif dari session tidak valid (misal dicabut admin) atau belum dipilih
		if len(currentBaseRoles) > 1 {
			// Jika user punya > 1 peran dasar SEKARANG, paksa pilih lagi
			session.Values["available_roles"] = currentBaseRoles // Update daftar peran di session
			delete(session.Values, "active_role")                // Hapus peran aktif yang lama
			session.Save(r, w)
			http.Redirect(w, r, "/select-role", http.StatusFound)
			return
		} else if len(currentBaseRoles) == 1 {
			// Jika user hanya punya 1 peran dasar SEKARANG, otomatis set itu
			finalActiveRole = currentBaseRoles[0]
			session.Values["active_role"] = finalActiveRole // Update session
			session.Save(r, w)                              // Simpan perubahan session
		} else {
			// Jika user tidak punya peran dasar SEKARANG (kasus aneh/dicabut semua)
			log.Printf("WARNING: User %d tidak memiliki peran dasar aktif.", userID)
			session.Options.MaxAge = -1 // Logout paksa
			session.Save(r, w)
			http.Error(w, "Anda tidak memiliki peran dasar yang aktif.", http.StatusForbidden)
			return
		}
	} else {
		// Jika peran aktif dari session masih valid, gunakan itu
		finalActiveRole = activeRoleFromSession
	}

	rolesToQuery := []string{finalActiveRole}
	if finalActiveRole == "dosen" {
		for _, attrMap := range currentAttributes { // currentAttributes didapat dari Fase 1
			if roleName, ok := attrMap["role"].(string); ok {
				rolesToQuery = append(rolesToQuery, roleName)
			}
		}
	}
	rolesToQuery = uniqueStrings(rolesToQuery)

	// Query aplikasi berdasarkan SEMUA peran TERBARU (base + attribute)
	var accessibleApps []Application // Atau []Application jika tidak perlu grouping
	if len(rolesToQuery) > 0 {
		queryArgs := make([]interface{}, len(rolesToQuery))
		for i, v := range rolesToQuery {
			queryArgs[i] = v
		}
		queryBase := `
			SELECT DISTINCT a.name, a.slug, a.target_url
			FROM applications a
			JOIN application_access aa ON a.id = aa.application_id
			JOIN roles r ON aa.role_id = r.id
			WHERE r.name IN (?)
            ORDER BY a.name`

		query, args, err := sqlx.In(queryBase, rolesToQuery)
		if err != nil {
			log.Printf("Error query applications for roles %v: %v", allCurrentRoleNames, err)
			http.Error(w, "Gagal mengambil data aplikasi", http.StatusInternalServerError)
			return
		}
		query = dc.env.DB.Rebind(query)

		appRows, err := dc.env.DB.Query(query, args...)

		if err != nil {
			log.Printf("Error query applications for roles %v: %v", rolesToQuery, err)
			http.Error(w, "Gagal mengambil data aplikasi", http.StatusInternalServerError)
			return
		}
		defer appRows.Close()

		for appRows.Next() {
			var app Application
			if err := appRows.Scan(&app.Name, &app.Slug, &app.TargetURL); err != nil {
				log.Printf("Error scanning application row: %v", err) // Log error scan
				continue
			}
			accessibleApps = append(accessibleApps, app)
		}
	}

	data := map[string]interface{}{
		"UserName":   user.Name,       // Gunakan nama terbaru dari DB
		"ActiveRole": finalActiveRole, // Gunakan peran aktif yang sudah divalidasi/diset
		"AppsByRole": accessibleApps,
	}

	// Render template dashboard
	err = dc.env.Templates.ExecuteTemplate(w, "dashboard.html", data)
	if err != nil {
		log.Printf("ERROR rendering dashboard template: %v", err)
		http.Error(w, "Gagal menampilkan halaman dashboard", http.StatusInternalServerError)
	}

}

func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

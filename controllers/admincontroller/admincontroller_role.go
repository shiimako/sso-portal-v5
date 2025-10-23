// file: controllers/admincontroller/admincontroller_role.go

package admincontroller

import (
	"log"
	"net/http"
	"sso-portal-v3/models"
)

// ListRoles menampilkan halaman daftar semua peran dan hak aksesnya.
func (ac *AdminController) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := models.GetAllRolesWithAccess(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data peran", http.StatusInternalServerError)
		log.Printf("ERROR: Gagal mengambil data peran: %v", err)
		return
	}

	data := map[string]interface{}{
		"Roles": roles,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-role-list.html", data)
}
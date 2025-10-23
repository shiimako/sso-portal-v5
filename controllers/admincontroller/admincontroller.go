// file: controllers/admincontroller/admincontroller.go

package admincontroller

import (
	"net/http"
	"sso-portal-v3/handlers" // Sesuaikan path
)

type AdminController struct {
	env *handlers.Env
}

func NewAdminController(env *handlers.Env) *AdminController {
	return &AdminController{env: env}
}

// Dashboard menampilkan halaman utama panel admin.
func (ac *AdminController) Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	// Kirim data nama admin ke template
	data := map[string]interface{}{
		"UserName": session.Values["user_name"],
	}

	// Render halaman dashboard admin
	err := ac.env.Templates.ExecuteTemplate(w, "admin-dashboard.html", data)
	if err != nil {
		http.Error(w, "Gagal render halaman admin", http.StatusInternalServerError)
	}
}

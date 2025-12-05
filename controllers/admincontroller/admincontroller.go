// file: controllers/admincontroller/admincontroller.go

package admincontroller

import (
	"net/http"
	"sso-portal-v3/config" // Sesuaikan path
	"sso-portal-v3/models"
	"sso-portal-v3/views"
)

type AdminController struct {
	env   *config.Env
	views *views.Views
}

func NewAdminController(env *config.Env, v *views.Views) *AdminController {
	return &AdminController{env: env, views: v}
}

// Dashboard menampilkan halaman utama panel admin.
func (ac *AdminController) Dashboard(w http.ResponseWriter, r *http.Request) {
	
	unreadErrors, err := models.CountUnreadErrors(ac.env.DB)
    if err != nil {
        // Jika gagal hitung (misal DB mati), set 0 saja biar dashboard tetap jalan
        unreadErrors = 0
    }

	pageData := map[string]interface{}{
		"UnreadErrors": unreadErrors,
	}

	ac.views.RenderPage(w, r, "admin-dashboard", pageData)
}

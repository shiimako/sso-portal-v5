// file: controllers/admincontroller/admincontroller.go

package admincontroller

import (
	"net/http"
	"sso-portal-v3/handlers" // Sesuaikan path
	"sso-portal-v3/views"
)

type AdminController struct {
	env *handlers.Env
	views *views.Views
}

func NewAdminController(env *handlers.Env, v *views.Views) *AdminController {
	return &AdminController{env: env, views:v,}
}

// Dashboard menampilkan halaman utama panel admin.
func (ac *AdminController) Dashboard(w http.ResponseWriter, r *http.Request) {

    pageData := map[string]interface{}{}

    ac.views.RenderPage(w, r, "admin_dashboard", pageData)
}


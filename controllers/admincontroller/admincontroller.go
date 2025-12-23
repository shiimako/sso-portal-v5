package admincontroller

import (
	"net/http"
	"sso-portal-v5/config"
	"sso-portal-v5/views"
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


	ac.views.RenderPage(w, r, "admin-dashboard", map[string]interface{}{})
}

func (ac *AdminController) RenderError(w http.ResponseWriter, r *http.Request, code int, message string) {
    w.WriteHeader(code)
    
    data := map[string]interface{}{
        "Code":    code,
        "Message": message,
    }
    
    ac.views.RenderPage(w, r, "error", data)
}

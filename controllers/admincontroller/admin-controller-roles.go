package admincontroller

import (
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (ac *AdminController) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data role", 500)
		return
	}
	ac.views.RenderPage(w, r, "admin-roles-list", map[string]interface{}{"Roles": roles})
}

func (ac *AdminController) NewRoleForm(w http.ResponseWriter, r *http.Request) {
	ac.views.RenderPage(w, r, "admin-roles-form", map[string]interface{}{"IsEdit": false})
}

func (ac *AdminController) CreateRole(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	desc := r.FormValue("description")

	if err := models.CreateRole(ac.env.DB, name, desc); err != nil {
		http.Error(w, "Gagal simpan role: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/admin/roles", http.StatusFound)
}

func (ac *AdminController) EditRoleForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	role, err := models.FindRoleByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "Role tidak ditemukan", 404)
		return
	}
	ac.views.RenderPage(w, r, "admin-roles-form", map[string]interface{}{
		"Role":   role,
		"IsEdit": true,
	})
}

func (ac *AdminController) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	name := r.FormValue("name")
	desc := r.FormValue("description")

	if err := models.UpdateRole(ac.env.DB, id, name, desc); err != nil {
		http.Error(w, "Gagal update role: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/admin/roles", http.StatusFound)
}

func (ac *AdminController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	models.DeleteRole(ac.env.DB, id)
	http.Redirect(w, r, "/admin/roles", http.StatusFound)
}

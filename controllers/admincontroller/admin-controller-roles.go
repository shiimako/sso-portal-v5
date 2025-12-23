package admincontroller

import (
	"database/sql"
	"log"
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (ac *AdminController) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal Mengambil data Roles")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	flashes := session.Flashes()
	_ = session.Save(r, w)

	ac.views.RenderPage(w, r, "admin-roles-list", map[string]interface{}{"Roles": roles, "Flash": flashes})
}

func (ac *AdminController) NewRoleForm(w http.ResponseWriter, r *http.Request) {
	ac.views.RenderPage(w, r, "admin-roles-form", map[string]interface{}{"IsEdit": false})
}

func (ac *AdminController) CreateRole(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	desc := r.FormValue("description")

	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama harus diisi")
		return
	}

	if err := models.CreateRole(ac.env.DB, name, desc); err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal simpan role")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Role berhasil ditambahkan.")
	session.Save(r, w)
	http.Redirect(w, r, "/admin/roles", http.StatusFound)
}

func (ac *AdminController) EditRoleForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	role, err := models.FindRoleByID(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Role Tidak Ditemukan")
			return
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal Mengambil data Role")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
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
	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama harus diisi")
		return
	}

	desc := r.FormValue("description")

	if err := models.UpdateRole(ac.env.DB, id, name, desc); err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal update role")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Role berhasil diupdate.")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/roles", http.StatusFound)
}

func (ac *AdminController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	err := models.DeleteRole(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Role Tidak Ditemukan")
			return
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal menghapus role")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Role berhasil dihapus.")
	session.Save(r, w)
	
	http.Redirect(w, r, "/admin/roles", http.StatusFound)
}

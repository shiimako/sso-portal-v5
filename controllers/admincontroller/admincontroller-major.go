package admincontroller

import (
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (ac *AdminController) ListMajors(w http.ResponseWriter, r *http.Request) {
	data, err := models.GetAllMajors(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengambil data jurusan.")
		return
	}
	ac.views.RenderPage(w, r, "admin-majors-list", map[string]interface{}{"Data": data})
}

func (ac *AdminController) NewMajorForm(w http.ResponseWriter, r *http.Request) {
	ac.views.RenderPage(w, r, "admin-majors-form", map[string]interface{}{
		"IsEdit": false,
	})
}

func (ac *AdminController) CreateMajor(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Nama jurusan tidak boleh kosong", http.StatusBadRequest)
		return
	}

	if err := models.CreateMajor(ac.env.DB, name); err != nil {
		http.Error(w, "Gagal menyimpan jurusan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/jurusan", http.StatusFound)
}

func (ac *AdminController) EditMajorForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	major, err := models.FindMajorByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "Jurusan tidak ditemukan", http.StatusNotFound)
		return
	}

	ac.views.RenderPage(w, r, "admin-majors-form", map[string]interface{}{
		"Major":  major,
		"IsEdit": true,
	})
}

func (ac *AdminController) UpdateMajor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if err := models.UpdateMajor(ac.env.DB, id, name); err != nil {
		http.Error(w, "Gagal update jurusan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/jurusan", http.StatusFound)
}

func (ac *AdminController) DeleteMajor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if err := models.DeleteMajor(ac.env.DB, id); err != nil {
		http.Error(w, "Gagal menghapus jurusan", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/jurusan", http.StatusFound)
}
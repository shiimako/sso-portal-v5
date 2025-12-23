package admincontroller

import (
	"database/sql"
	"log"
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

	sesssion, _ := ac.env.Store.Get(r, ac.env.SessionName)
	flashes := sesssion.Flashes()
	_ = sesssion.Save(r, w)

	ac.views.RenderPage(w, r, "admin-majors-list", map[string]interface{}{"Data": data, "Flash": flashes})
}

func (ac *AdminController) NewMajorForm(w http.ResponseWriter, r *http.Request) {
	ac.views.RenderPage(w, r, "admin-majors-form", map[string]interface{}{
		"IsEdit": false,
	})
}

func (ac *AdminController) CreateMajor(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Gagal Parsing Form")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama jurusan tidak boleh kosong")
		return
	}

	if err := models.CreateMajor(ac.env.DB, name); err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal menyimpan jurusan: "+err.Error())
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Jurusan berhasil ditambahkan.")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/jurusan", http.StatusFound)
}

func (ac *AdminController) EditMajorForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	major, err := models.FindMajorByID(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Jurusan Tidak Ditemukan")
			return
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengambil data jurusan.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
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
		ac.RenderError(w, r, http.StatusBadRequest, "Gagal Parsing Form")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama jurusan tidak boleh kosong")
		return
	}

	if err := models.UpdateMajor(ac.env.DB, id, name); err != nil {

		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Jurusan Tidak Ditemukan")
			return
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengupdate jurusan: "+err.Error())
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Jurusan berhasil diupdate.")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/jurusan", http.StatusFound)
}

func (ac *AdminController) DeleteMajor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if err := models.DeleteMajor(ac.env.DB, id); err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Jurusan Tidak Ditemukan")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal menghapus jurusan: "+err.Error())
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Jurusan berhasil dihapus.")
	session.Save(r, w)
	
	http.Redirect(w, r, "/admin/jurusan", http.StatusFound)
}
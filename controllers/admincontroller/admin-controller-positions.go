package admincontroller

import (
	"database/sql"
	"log"
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (ac *AdminController) ListPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := models.GetAllPositions(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	flashes := session.Flashes()
	_ = session.Save(r, w)
	ac.views.RenderPage(w, r, "admin-positions-list", map[string]interface{}{"Positions": positions, "Flash": flashes})
}

func (ac *AdminController) NewPositionForm(w http.ResponseWriter, r *http.Request) {
	ac.views.RenderPage(w, r, "admin-positions-form", map[string]interface{}{"IsEdit": false})
}

func (ac *AdminController) CreatePosition(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama jabatan tidak boleh kosong.")
		return
	}

	if err := models.CreatePosition(ac.env.DB, name); err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session , _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Jabatan berhasil ditambahkan.")
	session.Save(r, w)
	http.Redirect(w, r, "/admin/positions", http.StatusFound)
}

func (ac *AdminController) EditPositionForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	pos, err := models.FindPositionByID(ac.env.DB, id)
	if err != nil {

		if err == sql.ErrNoRows{
		ac.RenderError(w, r, http.StatusNotFound, "Data Posisi Tidak Ditemukan")
		return
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}
	ac.views.RenderPage(w, r, "admin-positions-form", map[string]interface{}{
		"Position": pos,
		"IsEdit":   true,
	})
}

func (ac *AdminController) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	name := r.FormValue("name")
	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama jabatan tidak boleh kosong.")
		return
	}

	if err := models.UpdatePosition(ac.env.DB, id, name); err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Posisi Tidak Ditemukan")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Jabatan berhasil diupdate.")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/positions", http.StatusFound)
}

func (ac *AdminController) DeletePosition(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	err := models.DeletePosition(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Jabatan Tidak Ditemukan")
			return
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Jabatan berhasil dihapus.")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/positions", http.StatusFound)
}
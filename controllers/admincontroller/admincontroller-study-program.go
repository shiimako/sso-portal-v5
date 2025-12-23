package admincontroller

import (
	"database/sql"
	"log"
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (ac *AdminController) ListStudyPrograms(w http.ResponseWriter, r *http.Request) {
	data, err := models.GetAllStudyPrograms(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengambil data prodi.")
		return
	}

	sessions, _ := ac.env.Store.Get(r, ac.env.SessionName)
	falshes := sessions.Flashes()
	_ = sessions.Save(r, w)

	ac.views.RenderPage(w, r, "admin-study-programs-list", map[string]interface{}{"Data": data, "Flash": falshes})
}

func (ac *AdminController) NewStudyProgramForm(w http.ResponseWriter, r *http.Request) {
	majors, err := models.GetAllMajors(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengambil data jurusan.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	ac.views.RenderPage(w, r, "admin-study-programs-form", map[string]interface{}{
		"Majors": majors,
		"IsEdit": false,
	})
}

func (ac *AdminController) CreateStudyProgram(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Gagal Parsing Form")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama Prodi tidak boleh kosong")
		return
	}

	majorID, _ := strconv.Atoi(r.FormValue("major_id"))

	if err := models.CreateStudyProgram(ac.env.DB, name, majorID); err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal menyimpan prodi: "+err.Error())
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	sessions, _ := ac.env.Store.Get(r, ac.env.SessionName)
	sessions.AddFlash("Prodi berhasil ditambahkan.")
	_ = sessions.Save(r, w)

	http.Redirect(w, r, "/admin/study-programs", http.StatusFound)
}

func (ac *AdminController) EditStudyProgramForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	prodi, err := models.FindStudyProgramByID(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Prodi Tidak Ditemukan")
			return
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengambil data prodi.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	majors, err := models.GetAllMajors(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengambil data jurusan.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	ac.views.RenderPage(w, r, "admin-study-programs-form", map[string]interface{}{
		"StudyProgram": prodi,
		"Majors":       majors,
		"IsEdit":       true,
	})
}

func (ac *AdminController) UpdateStudyProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if err := r.ParseForm(); err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Gagal Parsing Form")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama Prodi tidak boleh kosong")
		return
	}

	majorID, _ := strconv.Atoi(r.FormValue("major_id"))

	if err := models.UpdateStudyProgram(ac.env.DB, id, name, majorID); err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Prodi Tidak Ditemukan")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal mengupdate prodi: "+err.Error())
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Prodi berhasil diupdate.")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/study-programs", http.StatusFound)
}

func (ac *AdminController) DeleteStudyProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if err := models.DeleteStudyProgram(ac.env.DB, id); err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Prodi Tidak Ditemukan")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Gagal menghapus prodi: "+err.Error())
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Prodi berhasil dihapus.")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/study-programs", http.StatusFound)
}
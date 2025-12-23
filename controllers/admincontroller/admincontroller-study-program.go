package admincontroller

import (
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
	ac.views.RenderPage(w, r, "admin-study-programs-list", map[string]interface{}{"Data": data})
}

func (ac *AdminController) NewStudyProgramForm(w http.ResponseWriter, r *http.Request) {
	majors, _ := models.GetAllMajors(ac.env.DB)

	ac.views.RenderPage(w, r, "admin-study-programs-form", map[string]interface{}{
		"Majors": majors,
		"IsEdit": false,
	})
}

func (ac *AdminController) CreateStudyProgram(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	majorID, _ := strconv.Atoi(r.FormValue("major_id"))

	if err := models.CreateStudyProgram(ac.env.DB, name, majorID); err != nil {
		http.Error(w, "Gagal menyimpan prodi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/study-programs", http.StatusFound)
}

func (ac *AdminController) EditStudyProgramForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	prodi, err := models.FindStudyProgramByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "Prodi tidak ditemukan", http.StatusNotFound)
		return
	}
	majors, _ := models.GetAllMajors(ac.env.DB)

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
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	majorID, _ := strconv.Atoi(r.FormValue("major_id"))

	if err := models.UpdateStudyProgram(ac.env.DB, id, name, majorID); err != nil {
		http.Error(w, "Gagal update prodi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/study-programs", http.StatusFound)
}

func (ac *AdminController) DeleteStudyProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if err := models.DeleteStudyProgram(ac.env.DB, id); err != nil {
		http.Error(w, "Gagal menghapus prodi", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/study-programs", http.StatusFound)
}
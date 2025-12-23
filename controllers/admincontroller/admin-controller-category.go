package admincontroller

import (
	"database/sql"
	"log"
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (ac *AdminController) ListCategories(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	flashes := session.Flashes()
	session.Save(r, w)

	var flashMsg string
	if len(flashes) > 0 {
		flashMsg = flashes[0].(string)
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		p, _ := strconv.Atoi(pageStr)
		if p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		l, _ := strconv.Atoi(limitStr)
		if l > 0 {
			limit = l
		}
	}

	cats, err := models.ListCategories(ac.env.DB, page, limit)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	pageData := map[string]interface{}{
		"Cats":  cats,
		"Page":  page,
		"Limit": limit,
		"Flash": flashMsg,
	}

	ac.views.RenderPage(w, r, "admin-cats-list", pageData)
}

func (ac *AdminController) NewCategoriesForm(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{}

	ac.views.RenderPage(w, r, "admin-cats-form", data)
}

func (ac *AdminController) CreateCategory(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	max, err := models.GetMaxSort(ac.env.DB)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	sort := max + 1

	if name == "" {
		http.Error(w, "Nama harus diisi", http.StatusBadRequest)
		return
	}

	err = models.CreateCategory(ac.env.DB, name, sort)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Kategori " + name + " berhasil ditambahkan!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (ac *AdminController) EditCategoriesForm(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Kategori Tidak Ditemukan")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	cat, err := models.FindCategoryByID(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Kategori dengan ID tersebut tidak ditemukan.")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	ac.views.RenderPage(w, r, "admin-cats-edit", map[string]interface{}{
		"Cat": cat,
	})
}

func (ac *AdminController) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	name := r.FormValue("name")
	sortStr := r.FormValue("sort")

	sort, err := strconv.Atoi(sortStr)
	if err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Sort harus berupa Angka.")
	}

	if name == "" {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama harus diisi")
		return
	}

	exists, err := models.IsSortExists(ac.env.DB, sort)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	if exists {
		ac.RenderError(w, r, http.StatusBadRequest, "Urutan sudah digunakan")
		return
	}

	err = models.UpdateCategory(ac.env.DB, id, name, sort)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("Kategori berhasil diperbarui!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (ac *AdminController) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	err = models.DeleteCategory(ac.env.DB, id)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session.AddFlash("Kategori berhasil dihapus!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

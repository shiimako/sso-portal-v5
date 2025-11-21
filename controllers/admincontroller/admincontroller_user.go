package admincontroller

import (
	"fmt"
	"net/http"
	"sso-portal-v3/models"
	"strconv"

	"github.com/gorilla/mux"
)

// ListUsers menampilkan daftar semua pengguna.
func (ac *AdminController) ListUsers(w http.ResponseWriter, r *http.Request) {

	users, err := models.GetAllUsers(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data pengguna", http.StatusInternalServerError)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	flashes := session.Flashes()
	session.Save(r, w)

	data := map[string]interface{}{
		"Users":   users,
		"Flashes": flashes,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-user-list.html", data)
}

// DetailUser menampilkan halaman detail read-only untuk seorang pengguna.
func (ac *AdminController) DetailUser(w http.ResponseWriter, r *http.Request) {
	idstr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, "ID pengguna tidak valid", http.StatusBadRequest)
		return
	}

	user, err := models.FindUserByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal mengambil data pengguna: %v", err), http.StatusInternalServerError)
		return
	}
	role := user.Roles[0].Name

	var positions []string
	if role == "dosen" {
		positionsDetails, err := models.GetLecturerPositionsByLecturerID(ac.env.DB, user.Lecturer.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Gagal mengambil data posisi dosen: %v", err), http.StatusInternalServerError)
			return
		}

		for _, pos := range positionsDetails {
			if pos.Scopetype == "none" {
				positions = append(positions, pos.PositionName)
			} else {
				positions = append(positions, fmt.Sprintf("%s (%s: %s)", pos.PositionName, pos.Scopetype, pos.ScopeName.String))
			}
		}
	}

	data := map[string]interface{}{
		"User":  user,
		"Roles": role,
		"Positions": positions,
	}

	ac.env.Templates.ExecuteTemplate(w, "admin-user-detail.html", data)
}

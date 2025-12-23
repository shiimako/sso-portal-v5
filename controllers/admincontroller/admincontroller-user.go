package admincontroller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/gorilla/mux"
)

func (ac *AdminController) ListUsers(w http.ResponseWriter, r *http.Request) {

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")

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

	users, err := models.GetAllUsers(ac.env.DB, page, limit, search, role)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	pageData := map[string]interface{}{
		"Users":        users,
		"Page":         page,
		"Limit":        limit,
		"Search":       search,
		"Role":         role,
	}

	ac.views.RenderPage(w, r, "admin-user-list", pageData)
}

func (ac *AdminController) DetailUser(w http.ResponseWriter, r *http.Request) {
	idstr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idstr)
	if err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "ID Pengguna tidak Valid.")
		return
	}

	user, err := models.FindUserByID(ac.env.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ac.RenderError(w, r, http.StatusNotFound, "Data Pengguna Tidak Ditemukan")
			return
		}
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	profile := map[string]interface{}{
		"ID":      user.ID,
		"Name":    user.Name,
		"Email":   user.Email,
		"Avatar":  user.Avatar,
		"Address": user.Address,
		"Phone":   user.Phone,
		"NIM":     nil,
		"NIP":     nil,
		"NUPTK":   nil,
	}

	// Mahasiswa
	if user.Student != nil {
		profile["NIM"] = user.Student.NIM
	}

	// Dosen
	if user.Lecturer != nil {
		profile["NIP"] = user.Lecturer.NIP
		profile["NUPTK"] = user.Lecturer.NUPTK
	}

	role := user.Roles[0].Name

	var positions []string
	if role == "dosen" {
		positionsDetails, err := models.GetLecturerPositionsByLecturerID(ac.env.DB, user.Lecturer.ID)
		if err != nil {
			ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
			log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
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
		"User":      user,
		"Role":      role,
		"Positions": positions,
		"Profile":   profile,
	}

	ac.views.RenderPage(w, r, "admin-user-detail", data)
}

func (ac *AdminController) NewUserForm(w http.ResponseWriter, r *http.Request) {
	roles, _ := models.GetAllRoles(ac.env.DB)
	positions, _ := models.GetAllPositions(ac.env.DB)
	majors, _ := models.GetAllMajors(ac.env.DB)
	prodis, _ := models.GetAllStudyPrograms(ac.env.DB)

	ac.views.RenderPage(w, r, "admin-user-form", map[string]interface{}{
		"IsEdit":          false,
		"Roles":           roles,
		"MasterPositions": positions,
		"MasterMajors":    majors,
		"MasterProdis":    prodis,
	})
}

func (ac *AdminController) CreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parse form", 400)
		return
	}

	roleID, _ := strconv.Atoi(r.FormValue("role_id"))
	var roleName string
	ac.env.DB.Get(&roleName, "SELECT role_name FROM roles WHERE id = ?", roleID)

	form := models.UserForm{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Status:   r.FormValue("status"),
		RoleID:   roleID,
		RoleName: roleName,
		Address:  models.GetPtr(r.FormValue("address")),
		Phone:    models.GetPtr(r.FormValue("phone")),
		NIM:      models.GetPtr(r.FormValue("nim")),
		NIP:      models.GetPtr(r.FormValue("nip")),
		NUPTK:    models.GetPtr(r.FormValue("nuptk")),
	}
    if form.Status == "" { form.Status = "active" }

	positions, err := models.ParseLecturerPositions(r.FormValue("positions_json"))
	if err != nil {
		http.Error(w, "Format Jabatan Error", 400)
		return
	}
	form.Positions = positions

	_, err = models.CreateUser(ac.env.DB, form)
	if err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Email sudah digunakan!!")
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (ac *AdminController) EditUserForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	user, err := models.FindUserByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, "User not found", 404)
		return
	}

	roles, _ := models.GetAllRoles(ac.env.DB)
	positions, _ := models.GetAllPositions(ac.env.DB)
	majors, _ := models.GetAllMajors(ac.env.DB)
	prodis, _ := models.GetAllStudyPrograms(ac.env.DB)


	var existingPosJSON []models.PositionFormJSON
	if user.Positions != nil {
		for _, p := range user.Positions {
			item := models.PositionFormJSON{
				PositionID: p.PositionID,
				Scope:      "none",
			}
            
			if p.StartDate != nil {
				item.StartDate = *p.StartDate
			}
			if p.EndDate != nil {
				item.EndDate = *p.EndDate
			}

			if p.MajorID.Valid {
				item.Scope = "major"
				item.MajorID = int(p.MajorID.Int64)
			} else if p.StudyProgramID.Valid {
				item.Scope = "prodi"
				item.ProdiID = int(p.StudyProgramID.Int64)
			}
			existingPosJSON = append(existingPosJSON, item)
		}
	}
	posBytes, _ := json.Marshal(existingPosJSON)
    if len(existingPosJSON) == 0 { posBytes = []byte("[]") }

	ac.views.RenderPage(w, r, "admin-user-form", map[string]interface{}{
		"IsEdit":           true,
		"User":             user,
		"Roles":            roles,
		"MasterPositions":  positions,
		"MasterMajors":     majors,
		"MasterProdis":     prodis,
		"CurrentPositions": string(posBytes),
	})
}

func (ac *AdminController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	r.ParseForm()

	roleID, _ := strconv.Atoi(r.FormValue("role_id"))
	var roleName string
	ac.env.DB.Get(&roleName, "SELECT role_name FROM roles WHERE id = ?", roleID)

	form := models.UserForm{
		ID:       id,
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Status:   r.FormValue("status"),
		RoleID:   roleID,
		RoleName: roleName,
		Address:  models.GetPtr(r.FormValue("address")),
		Phone:    models.GetPtr(r.FormValue("phone")),
		NIM:      models.GetPtr(r.FormValue("nim")),
		NIP:      models.GetPtr(r.FormValue("nip")),
		NUPTK:    models.GetPtr(r.FormValue("nuptk")),
	}

	positions, err := models.ParseLecturerPositions(r.FormValue("positions_json"))
	if err != nil {
		http.Error(w, "Format Jabatan Error", 400)
		return
	}
	form.Positions = positions

	err = models.UpdateUser(ac.env.DB, form)
	if err != nil {
		http.Error(w, "Gagal update: "+err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (ac *AdminController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := models.DeleteUser(ac.env.DB, id)
	if err != nil {
		fmt.Printf("Error Delete User: %v\n", err)
		http.Error(w, "Gagal menghapus user", 500)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

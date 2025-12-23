package admincontroller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sso-portal-v5/models"
	"strconv"

	"github.com/go-sql-driver/mysql"
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

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	flashes := session.Flashes()
	session.Save(r, w)

	pageData := map[string]interface{}{
		"Users":  users,
		"Page":   page,
		"Limit":  limit,
		"Search": search,
		"Role":   role,
		"Flash":  flashes,
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
	if form.Status == "" {
		form.Status = "active"
	}

	if form.Name == "" || form.Email == "" || roleID == 0 {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama, Email, dan Role wajib diisi!")
		return
	}

	if form.RoleName == "mahasiswa" {
		if r.FormValue("nim") == "" {
			ac.RenderError(w, r, http.StatusBadRequest, "NIM Wajib diisi untuk Mahasiswa!")
			return
		}
	} else if form.RoleName == "dosen" {
		if r.FormValue("nip") == "" || r.FormValue("nuptk") == "" {
			ac.RenderError(w, r, http.StatusBadRequest, "Dosen wajib memiliki NIP atau NUPTK!")
			return
		}
	}

	positions, err := models.ParseLecturerPositions(r.FormValue("positions_json"))
	if err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Parsing Position Error")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}
	form.Positions = positions

	_, err = models.CreateUser(ac.env.DB, form)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				ac.RenderError(w, r, http.StatusBadRequest, "Email sudah digunakan! Silahkan gunakan email lain.")
				return
			}
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("User " + form.Name + " berhasil ditambahkan!")
	session.Save(r, w)

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
	if len(existingPosJSON) == 0 {
		posBytes = []byte("[]")
	}

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

	if form.Name == "" || form.Email == "" || roleID == 0 {
		ac.RenderError(w, r, http.StatusBadRequest, "Nama, Email, dan Role wajib diisi!")
		return
	}

	if form.RoleName == "mahasiswa" {
		if r.FormValue("nim") == "" {
			ac.RenderError(w, r, http.StatusBadRequest, "NIM Wajib diisi untuk Mahasiswa!")
			return
		}
	} else if form.RoleName == "dosen" {
		if r.FormValue("nip") == "" || r.FormValue("nuptk") == "" {
			ac.RenderError(w, r, http.StatusBadRequest, "Dosen wajib memiliki NIP atau NUPTK!")
			return
		}
	}

	positions, err := models.ParseLecturerPositions(r.FormValue("positions_json"))
	if err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Parsing Position Error")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}
	form.Positions = positions

	err = models.UpdateUser(ac.env.DB, form)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				ac.RenderError(w, r, http.StatusBadRequest, "Email sudah digunakan! Silahkan gunakan email lain.")
				return
			}
		}

		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("User " + form.Name + " berhasil diubah!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (ac *AdminController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	loggedInUser := r.Context().Value("UserLogin").(*models.FullUser)
    
    if loggedInUser.ID == id {
        session, _ := ac.env.Store.Get(r, ac.env.SessionName)
        session.AddFlash("Anda tidak bisa menghapus akun Anda sendiri!")
        session.Save(r, w)
        http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
        return
    }

	err := models.DeleteUser(ac.env.DB, id)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.AddFlash("User dengan ID " + vars["id"] + " berhasil dihapus!")
	session.Save(r, w)

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

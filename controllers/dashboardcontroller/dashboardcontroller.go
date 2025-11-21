// file: controllers/dashboardcontroller/dashboardcontroller.go

package dashboardcontroller

import (
	"fmt"
	"log"
	"net/http"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"
)

type DashboardController struct {
	env *handlers.Env
}

func NewDashboardController(env *handlers.Env) *DashboardController {
	return &DashboardController{env: env}
}

func (dc *DashboardController) Index(w http.ResponseWriter, r *http.Request) {
	session, _ := dc.env.Store.Get(r, dc.env.SessionName)

	userID, ok := session.Values["user_id"]
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user, err := models.FindUserByID(dc.env.DB, userID.(int))
	if err != nil {
		log.Printf("ERROR: Gagal mengambil user %d dari DB: %v", userID, err)
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Error(w, "Gagal memuat data pengguna.", http.StatusInternalServerError)
		return
	}

	if user.Status != "aktif" {
		log.Printf("INFO: Akses dashboard ditolak untuk user %d karena status: %s", userID, user.Status)
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Error(w, fmt.Sprintf("Akun Anda berstatus '%s'. Tidak dapat mengakses dashboard.", user.Status), http.StatusForbidden)
		return
	}

	role := user.Roles[0].Name

	activeRole, ok := session.Values["active_role"].(string)
	isValid := ok && activeRole == role

	var finalActiveRole string
	finalActiveRole = role
	if isValid {
		finalActiveRole = activeRole
	} else {
		session.Values["active_role"] = role
		session.Save(r, w)
	}

	positionsToQuery := []int{}

	if finalActiveRole == "dosen" {
		for _, pos := range user.Positions {
			positionsToQuery = append(positionsToQuery, pos.PositionID)
		}
	}

	roleapps, err := models.FindApplicationsByRole(dc.env.DB, finalActiveRole)
	if err != nil {
    log.Printf("ERROR: Gagal mengambil aplikasi untuk role %s user %d: %v", finalActiveRole, user.ID, err)
    http.Error(w, "Gagal memuat aplikasi untuk dashboard.", http.StatusInternalServerError)
    return
	}

	positionapps := []models.Application{}
	if finalActiveRole == "dosen" && len(positionsToQuery) > 0 {
		positionapps, err = models.FindApplicationsByPositions(dc.env.DB, positionsToQuery)
		if err != nil {
			log.Printf("ERROR: Gagal mengambil aplikasi untuk posisi dosen user %d: %v", user.ID, err)
			http.Error(w, "Gagal memuat aplikasi untuk dashboard.", http.StatusInternalServerError)
			return
		}
	}
	apps := map[int]models.Application{}
	for _, app := range roleapps {
		apps[app.ID] = app
	}
	for _, app := range positionapps {
		apps[app.ID] = app
	}

	data := map[string]interface{}{
		"UserName":   user.Name,
		"ActiveRole": finalActiveRole,
		"Apps": apps,
	}

	err = dc.env.Templates.ExecuteTemplate(w, "dashboard.html", data)
	if err != nil {
		log.Printf("ERROR rendering dashboard template: %v", err)
		http.Error(w, "Gagal menampilkan halaman dashboard", http.StatusInternalServerError)
	}

}

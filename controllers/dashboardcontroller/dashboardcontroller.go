// file: controllers/dashboardcontroller/dashboardcontroller.go

package dashboardcontroller

import (
	"log"
	"net/http"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"
	"sso-portal-v3/views"
)

type DashboardController struct {
	env *handlers.Env
	views *views.Views
}

func NewDashboardController(env *handlers.Env, v *views.Views) *DashboardController {
	return &DashboardController{
		env: env,
		views: v,
	}
}

func (dc *DashboardController) Index(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("UserLogin").(*models.FullUser)
	role := user.Roles[0].Name

	positionsToQuery := []int{}

	if role == "dosen" {
		for _, pos := range user.Positions {
			positionsToQuery = append(positionsToQuery, pos.PositionID)
		}
	}

	roleapps, err := models.FindApplicationsByRole(dc.env.DB, role)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil aplikasi untuk role %s user %d: %v", role, user.ID, err)
		http.Error(w, "Gagal memuat aplikasi untuk dashboard.", http.StatusInternalServerError)
		return
	}

	positionapps := []models.Application{}
	if role == "dosen" && len(positionsToQuery) > 0 {
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

	adminContact, err := models.GetAdminContact(dc.env.DB)
	if err != nil {
		log.Println("ERROR: Gagal mengambil contact admin : ", err)
		http.Error(w, "Gagal memuat kontak admin.", http.StatusInternalServerError)
		return
	}
	
	dc.views.RenderPage(w, r, "dashboard", map[string]interface{}{
        "Apps":  apps,
        "Admin": adminContact,
    })

}

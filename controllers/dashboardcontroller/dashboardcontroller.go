// file: controllers/dashboardcontroller/dashboardcontroller.go

package dashboardcontroller

import (
	"log"
	"net/http"
	"sso-portal-v3/config"
	"sso-portal-v3/models"
	"sso-portal-v3/views"
)

type DashboardController struct {
	env   *config.Env
	views *views.Views
}

func NewDashboardController(env *config.Env, v *views.Views) *DashboardController {
	return &DashboardController{
		env:   env,
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

	apps, err := models.FindAccessibleApps(dc.env.DB, role, positionsToQuery)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil aplikasi untuk role %s user %d: %v", role, user.ID, err)
		http.Error(w, "Gagal memuat aplikasi untuk dashboard.", http.StatusInternalServerError)
		return
	}

	adminContact, err := models.GetContact(dc.env.DB, dc.env.AdminEmail)
	if err != nil {
		log.Println("ERROR: Gagal mengambil contact admin : ", err)
		http.Error(w, "Gagal memuat kontak admin.", http.StatusInternalServerError)
		return
	}

	unreadErrors, err := models.CountUnreadErrors(dc.env.DB)
    if err != nil {
        unreadErrors = 0
    }

	dc.views.RenderPage(w, r, "dashboard", map[string]interface{}{
		"Apps":  apps,
		"Admin": adminContact,
		"UnreadErrors": unreadErrors,
	})

}

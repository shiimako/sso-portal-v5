// file: controllers/dashboardcontroller/dashboardcontroller.go

package dashboardcontroller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sso-portal-v5/config"
	"sso-portal-v5/models"
	"sso-portal-v5/views"
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

	allCategories, err := models.GetAllCategories(dc.env.DB)
	if err != nil {
		dc.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	catIDStr := r.URL.Query().Get("cat_id")
	activeCatID := 0

	if catIDStr != "" {
        fmt.Sscanf(catIDStr, "%d", &activeCatID)
    } else if len(allCategories) > 0 {
        activeCatID = allCategories[0].ID
    }

	apps, err := models.FindAccessibleApps(dc.env.DB, role, positionsToQuery, activeCatID)
	if err != nil {
		dc.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	adminContact, err := models.GetContact(dc.env.DB, dc.env.AdminEmail)
	if err != nil {
		dc.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	notifications, err := models.GetNotificationSummary(dc.env.DB, user.ID)
	if err != nil {
        log.Println("WARNING: Gagal mengambil notifikasi user:", err)
        notifications = make(map[string]models.NotifSummary)
    }

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")

		if apps == nil {
			apps = []models.Application{} 
		}
		
		if notifications == nil {
			notifications = make(map[string]models.NotifSummary)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"Apps":   apps,
			"Notifs": notifications,
		})
		return 
	}

	vapidpublickey := dc.env.VapidPublicKey

	dc.views.RenderPage(w, r, "dashboard", map[string]interface{}{
		"Apps":  apps,
		"Admin": adminContact,
		"Notifs": notifications,
		"VapidPublicKey": vapidpublickey,
		"Categories": allCategories,
		"ActiveCatID":    activeCatID,
	})

}

func (ac *DashboardController) RenderError(w http.ResponseWriter, r *http.Request, code int, message string) {
    w.WriteHeader(code)
    
    data := map[string]interface{}{
        "Code":    code,
        "Message": message,
    }
    
    ac.views.RenderPage(w, r, "error", data)
}

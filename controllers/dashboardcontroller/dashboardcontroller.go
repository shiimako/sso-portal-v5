// file: controllers/dashboardcontroller/dashboardcontroller.go

package dashboardcontroller

import (
	"fmt"
	"log"
	"net/http"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"

	"github.com/jmoiron/sqlx"
)

type Application struct {
	Name      string
	Slug      string
	TargetURL string
}

type DashboardController struct {
	env *handlers.Env
}

func NewDashboardController(env *handlers.Env) *DashboardController {
	return &DashboardController{env: env}
}

func (dc *DashboardController) Index(w http.ResponseWriter, r *http.Request) {
	session, _ := dc.env.Store.Get(r, dc.env.SessionName)

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user, err := models.FindUserByID(dc.env.DB, fmt.Sprintf("%d", userID))
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

	allRolesFromDB, err := models.GetUserRolesAndAttributes(dc.env.DB, userID)
	if err != nil {
		log.Printf("ERROR: Gagal mengambil peran terbaru untuk user %d: %v", userID, err)
		http.Error(w, "Gagal mengambil data peran.", http.StatusInternalServerError)
		return
	}

	var currentBaseRoles []string
	var currentAttributes []map[string]interface{}
	var allCurrentRoleNames []string

	for _, role := range allRolesFromDB {
		allCurrentRoleNames = append(allCurrentRoleNames, role.Name)
		if role.Type == "base" {
			currentBaseRoles = append(currentBaseRoles, role.Name)
		} else if role.Type == "attribute" {
			attr := map[string]interface{}{"role": role.Name}
			if role.Scope.Valid {
				attr["scope"] = role.Scope.String
			}
			currentAttributes = append(currentAttributes, attr)
		}
	}

	activeRoleFromSession, sessionHasActiveRole := session.Values["active_role"].(string)
	isValidActiveRole := false
	if sessionHasActiveRole && activeRoleFromSession != "" {
		for _, dbRole := range currentBaseRoles {
			if dbRole == activeRoleFromSession {
				isValidActiveRole = true
				break
			}
		}
	}

	var finalActiveRole string

	if !isValidActiveRole {
		if len(currentBaseRoles) == 1 {
			finalActiveRole = currentBaseRoles[0]
			session.Values["active_role"] = finalActiveRole
			session.Save(r, w)
		} else {
			log.Printf("WARNING: User %d tidak memiliki peran dasar aktif.", userID)
			session.Options.MaxAge = -1
			session.Save(r, w)
			http.Error(w, "Anda tidak memiliki peran dasar yang aktif.", http.StatusForbidden)
			return
		}
	} else {
		finalActiveRole = activeRoleFromSession
	}

	rolesToQuery := []string{finalActiveRole}
	if finalActiveRole == "dosen" {
		for _, attrMap := range currentAttributes {
			if roleName, ok := attrMap["role"].(string); ok {
				rolesToQuery = append(rolesToQuery, roleName)
			}
		}
	}
	rolesToQuery = uniqueStrings(rolesToQuery)

	appsByRole := make(map[string][]Application)
	if len(rolesToQuery) > 0 {
		queryArgs := make([]interface{}, len(rolesToQuery))
		for i, v := range rolesToQuery {
			queryArgs[i] = v
		}
		queryBase := `
			SELECT a.name, a.slug, a.target_url, r.name as granting_role
			FROM applications a
			JOIN application_access aa ON a.id = aa.application_id
			JOIN roles r ON aa.role_id = r.id
			WHERE r.name IN (?)
            ORDER BY r.name, a.name`

		query, args, err := sqlx.In(queryBase, rolesToQuery)
		if err != nil {
			log.Printf("Error query applications for roles %v: %v", allCurrentRoleNames, err)
			http.Error(w, "Gagal mengambil data aplikasi", http.StatusInternalServerError)
			return
		}
		query = dc.env.DB.Rebind(query)

		appRows, err := dc.env.DB.Query(query, args...)

		if err != nil {
			log.Printf("Error query applications for roles %v: %v", rolesToQuery, err)
			http.Error(w, "Gagal mengambil data aplikasi", http.StatusInternalServerError)
			return
		}
		defer appRows.Close()

		for appRows.Next() {
			var app Application
			var grantingRole string
			if err := appRows.Scan(&app.Name, &app.Slug, &app.TargetURL, &grantingRole); err != nil {
				log.Printf("Error scanning application row: %v", err)
				continue
			}
			appsByRole[grantingRole] = append(appsByRole[grantingRole], app)
		}
	}

	data := map[string]interface{}{
		"UserName":   user.Name,
		"ActiveRole": finalActiveRole,
		"AppsByRole": appsByRole,
	}

	err = dc.env.Templates.ExecuteTemplate(w, "dashboard.html", data)
	if err != nil {
		log.Printf("ERROR rendering dashboard template: %v", err)
		http.Error(w, "Gagal menampilkan halaman dashboard", http.StatusInternalServerError)
	}

}

func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

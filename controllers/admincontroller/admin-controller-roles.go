package admincontroller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sso-portal-v3/models"
	"sso-portal-v3/services"
)

func (ac *AdminController) ListRoles(w http.ResponseWriter, r *http.Request) {
	data, err := models.GetAllRoles(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data role", 500)
		fmt.Println("ERROR : ", err)
		return
	}
	ac.views.RenderPage(w, r, "admin-roles-list", map[string]interface{}{"Data": data})
}

// StreamRoleSync menangani SSE untuk Role
func (ac *AdminController) StreamRoleSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", 500)
		return
	}

	sendJSON := func(data map[string]interface{}) {
		jsonMsg, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonMsg)
		flusher.Flush()
	}

	models.CreateLog(ac.env.DB, "MANUAL", "ROLE", "RUNNING", "Memulai sync Role.")

	serviceReporter := func(progress int, msg string) {
		sendJSON(map[string]interface{}{"progress": progress, "log": msg, "status": "running"})
	}

	err := services.SyncRoles(ac.env.DB, serviceReporter)

	if err != nil {
		models.CreateLog(ac.env.DB, "MANUAL", "ROLE", "ERROR", err.Error())
		sendJSON(map[string]interface{}{"status": "error", "message": err.Error(), "log": "❌ Gagal Sync."})
	} else {
		models.CreateLog(ac.env.DB, "MANUAL", "ROLE", "SUCCESS", "Sinkronisasi Role Berhasil.")
		sendJSON(map[string]interface{}{"status": "done", "log": "✨ Selesai."})
	}
}
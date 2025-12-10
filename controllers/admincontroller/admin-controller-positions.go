package admincontroller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sso-portal-v3/models"
	"sso-portal-v3/services"
	"strings"
)

func (ac *AdminController) ListPositions(w http.ResponseWriter, r *http.Request) {
	data, err := models.GetAllPositions(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data jabatan", 500)
		return
	}
	ac.views.RenderPage(w, r, "admin-positions-list", map[string]interface{}{"Data": data})
}

// StreamJabatanSync menangani SSE untuk Jabatan
func (ac *AdminController) StreamPositionsSync(w http.ResponseWriter, r *http.Request) {
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

	models.CreateLog(ac.env.DB, "MANUAL", "JABATAN", "RUNNING", "Memulai sync Jabatan.")

	serviceReporter := func(progress int, msg string) {
		sendJSON(map[string]interface{}{"progress": progress, "log": msg, "status": "running"})
	}

	err := services.SyncPositions(ac.env, serviceReporter, "")

	if err != nil {
		models.CreateLog(ac.env.DB, "MANUAL", "JABATAN", "ERROR", err.Error())
		sendJSON(map[string]interface{}{"status": "error", "message": err.Error(), "log": "❌ Gagal Sync."})
	} else {
		models.CreateLog(ac.env.DB, "MANUAL", "JABATAN", "SUCCESS", "Sinkronisasi Jabatan Berhasil.")
		sendJSON(map[string]interface{}{"status": "done", "log": "✨ Selesai."})
	}
}

func (ac *AdminController) RunPositionsCron() {

	log.Println("⏰ [CRON] Memulai Sync Jabatan...")
	models.CreateLog(ac.env.DB, "CRON", "JABATAN", "RUNNING", "Cron job jabatan berjalan otomatis.")

	lastTime, _ := models.GetLastSuccessTime(ac.env.DB, "JABATAN")

	cronReporter := func(progress int, msg string) {
		if progress == 100 || strings.Contains(msg, "❌") {
			log.Printf("[CRON JABATAN] %s", msg)
		}
	}

	err := services.SyncPositions(ac.env, cronReporter, lastTime)
	if err != nil {
		log.Printf("❌ [CRON] Jabatan Gagal: %v", err)
		models.CreateLog(ac.env.DB, "CRON", "JABATAN", "ERROR", err.Error())
	} else {
		log.Println("✅ [CRON] Jabatan Selesai.")
		models.CreateLog(ac.env.DB, "CRON", "JABATAN", "SUCCESS", "Cron job role berhasil selesai.")
	}
}

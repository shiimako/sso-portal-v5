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

func (ac *AdminController) ListStudyPrograms(w http.ResponseWriter, r *http.Request) {
	data, err := models.GetAllStudyPrograms(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data prodi", 500)
		return
	}
	ac.views.RenderPage(w, r, "admin-study-programs-list", map[string]interface{}{"Data": data})
}

func (ac *AdminController) StreamStudyProgramsSync(w http.ResponseWriter, r *http.Request) {
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


	models.CreateLog(ac.env.DB, "MANUAL", "PRODI", "RUNNING", "Memulai sync Prodi.")


	serviceReporter := func(progress int, msg string) {
		sendJSON(map[string]interface{}{"progress": progress, "log": msg, "status": "running"})
	}


	err := services.SyncStudyPrograms(ac.env, serviceReporter, "")


	if err != nil {
		models.CreateLog(ac.env.DB, "MANUAL", "PRODI", "ERROR", err.Error())
		sendJSON(map[string]interface{}{"status": "error", "message": err.Error(), "log": "❌ Gagal Sync."})
	} else {
		models.CreateLog(ac.env.DB, "MANUAL", "PRODI", "SUCCESS", "Sinkronisasi Prodi Berhasil.")
		sendJSON(map[string]interface{}{"status": "done", "log": "✨ Selesai."})
	}
}

func (ac *AdminController) RunStudyProgramsCron() {
	
	log.Println("⏰ [CRON] Memulai Sync Prodi...")
	models.CreateLog(ac.env.DB, "CRON", "PRODI", "RUNNING", "Cron job jurusan berjalan otomatis.")

	lastTime, _ := models.GetLastSuccessTime(ac.env.DB, "PRODI")

	cronReporter := func(progress int, msg string) {
		if progress == 100 || strings.Contains(msg, "❌") {
			log.Printf("[CRON PRODI] %s", msg)
		}
	}

	err := services.SyncStudyPrograms(ac.env, cronReporter, lastTime)

	if err != nil {
		log.Printf("❌ [CRON] Prodi Gagal: %v", err)
		models.CreateLog(ac.env.DB, "CRON", "PRODI", "ERROR", err.Error())
	} else {
		log.Println("✅ [CRON] Prodi Selesai.")
		models.CreateLog(ac.env.DB, "CRON", "PRODI", "SUCCESS", "Cron job Prodi berhasil selesai.")
	}
}
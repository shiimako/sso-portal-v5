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


func (ac *AdminController) ListMajors(w http.ResponseWriter, r *http.Request) {
	data, err := models.GetAllMajors(ac.env.DB)
	if err != nil {
		http.Error(w, "Gagal mengambil data jurusan", 500)
		return
	}
	ac.views.RenderPage(w, r, "admin-majors-list", map[string]interface{}{"Data": data})
}

func (ac *AdminController) StreamMajorsSync(w http.ResponseWriter, r *http.Request) {
	// Setup Header SSE
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

	models.CreateLog(ac.env.DB, "MANUAL", "JURUSAN", "RUNNING", "Memulai sinkronisasi Jurusan.")

	serviceReporter := func(progress int, msg string) {
		sendJSON(map[string]interface{}{
			"progress": progress,
			"log":      msg,
			"status":   "running",
		})
	}

	err := services.SyncMajors(ac.env, serviceReporter, "")

	if err != nil {
		models.CreateLog(ac.env.DB, "MANUAL", "JURUSAN", "ERROR", err.Error())
		sendJSON(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
			"log":     "❌ Gagal Sync.",
		})
	} else {
		models.CreateLog(ac.env.DB, "MANUAL", "JURUSAN", "SUCCESS", "Sinkronisasi Jurusan Berhasil.")
		sendJSON(map[string]interface{}{
			"status": "done",
			"log":    "✅ Selesai.",
		})
	}
}

func (ac *AdminController) RunMajorsCron() {
	
	log.Println("⏰ [CRON] Memulai Sync Jurusan...")

	// 1. Catat Log DB: START
	models.CreateLog(ac.env.DB, "CRON", "JURUSAN", "RUNNING", "Cron job jurusan berjalan otomatis.")

	lastTime, _ := models.GetLastSuccessTime(ac.env.DB, "JURUSAN")

	// 2. Buat Reporter "Bisu" (Cuma print ke console server, bukan HTTP)
	cronReporter := func(progress int, msg string) {
		// Kita bisa filter, misal cuma lapor kalau 100% atau Error
		if progress == 100 || strings.Contains(msg, "❌") {
			log.Printf("[CRON JURUSAN] %s", msg)
		}
	}

	// 3. Eksekusi Service
	err := services.SyncMajors(ac.env, cronReporter, lastTime)

	// 4. Catat Log DB: FINISH
	if err != nil {
		log.Printf("❌ [CRON] Jurusan Gagal: %v", err)
		models.CreateLog(ac.env.DB, "CRON", "JURUSAN", "ERROR", err.Error())
	} else {
		log.Println("✅ [CRON] Jurusan Selesai.")
		models.CreateLog(ac.env.DB, "CRON", "JURUSAN", "SUCCESS", "Cron job jurusan berhasil selesai.")
	}
}
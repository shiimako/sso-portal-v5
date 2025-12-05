package admincontroller

import (
	"net/http"
	"sso-portal-v3/models"
)

// SyncLogsPage menampilkan halaman riwayat log
func (ac *AdminController) SyncLogsPage(w http.ResponseWriter, r *http.Request) {
    
    // 1. Mark as Read (Karena admin membuka halaman ini, anggap sudah baca)
    go models.MarkErrorsAsRead(ac.env.DB)

    // 2. Ambil data logs
    limit := 50
    offset := 0
    logs, err := models.GetLogs(ac.env.DB, limit, offset)
    if err != nil {
        http.Error(w, "Gagal mengambil log", 500)
        return
    }

    pageData := map[string]interface{}{
        "Logs": logs,
    }

    ac.views.RenderPage(w, r, "admin-sync-logs", pageData)
}
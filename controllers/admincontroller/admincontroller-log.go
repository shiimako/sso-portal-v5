package admincontroller

import (
	"math"
	"net/http"
	"sso-portal-v3/models"
	"strconv"
)

// SyncLogsPage menampilkan halaman riwayat log dengan Pagination
func (ac *AdminController) SyncLogsPage(w http.ResponseWriter, r *http.Request) {
    
    // 1. Mark Read
    go models.MarkErrorsAsRead(ac.env.DB)

    // 2. Ambil Params
    pageStr := r.URL.Query().Get("page")
    moduleFilter := r.URL.Query().Get("module") // <--- Ambil filter

    page, _ := strconv.Atoi(pageStr)
    if page < 1 { page = 1 }
    limit := 20
    offset := (page - 1) * limit

    // 3. Ambil Data (Pass moduleFilter)
    logs, err := models.GetLogs(ac.env.DB, limit, offset, moduleFilter)
    if err != nil {
        logs = []models.SyncLog{}
    }

    // 4. Hitung Total (Pass moduleFilter)
    totalLogs, _ := models.CountLogs(ac.env.DB, moduleFilter)
    totalPages := int(math.Ceil(float64(totalLogs) / float64(limit)))

    // 5. Data View
    pageData := map[string]interface{}{
        "Logs":        logs,
        "CurrentPage": page,
        "TotalPages":  totalPages,
        "NextPage":    page + 1,
        "PrevPage":    page - 1,
        "TotalData":   totalLogs,
        "Module":      moduleFilter, // Kirim balik ke view biar dropdown terpilih
    }

    ac.views.RenderPage(w, r, "admin-sync-logs", pageData)
}
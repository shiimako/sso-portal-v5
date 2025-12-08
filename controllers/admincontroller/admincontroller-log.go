package admincontroller

import (
	"math"
	"net/http"
	"sso-portal-v3/models"
	"strconv"
)

// SyncLogsPage menampilkan halaman riwayat log dengan Pagination
func (ac *AdminController) SyncLogsPage(w http.ResponseWriter, r *http.Request) {
    
    go models.MarkErrorsAsRead(ac.env.DB)


    pageStr := r.URL.Query().Get("page")
    moduleFilter := r.URL.Query().Get("module") 

    page, _ := strconv.Atoi(pageStr)
    if page < 1 { page = 1 }
    limit := 20
    offset := (page - 1) * limit

    logs, err := models.GetLogs(ac.env.DB, limit, offset, moduleFilter)
    if err != nil {
        logs = []models.SyncLog{}
    }

    totalLogs, _ := models.CountLogs(ac.env.DB, moduleFilter)
    totalPages := int(math.Ceil(float64(totalLogs) / float64(limit)))


    pageData := map[string]interface{}{
        "Logs":        logs,
        "CurrentPage": page,
        "TotalPages":  totalPages,
        "NextPage":    page + 1,
        "PrevPage":    page - 1,
        "TotalData":   totalLogs,
        "Module":      moduleFilter, 
    }

    ac.views.RenderPage(w, r, "admin-sync-logs", pageData)
}
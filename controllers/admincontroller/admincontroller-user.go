package admincontroller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sso-portal-v3/models"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ListUsers menampilkan daftar semua pengguna.
func (ac *AdminController) ListUsers(w http.ResponseWriter, r *http.Request) {

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")

	page := 1
	limit := 10

	if pageStr != "" {
		p, _ := strconv.Atoi(pageStr)
		if p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		l, _ := strconv.Atoi(limitStr)
		if l > 0 {
			limit = l
		}
	}

	users, err := models.GetAllUsers(ac.env.DB, page, limit, search, role)
	if err != nil {
		http.Error(w, "Gagal mengambil data pengguna", http.StatusInternalServerError)
		return
	}

	unreadErrors, _ := models.CountUnreadErrors(ac.env.DB)

	pageData := map[string]interface{}{
		"Users": users,
		"Page" : page,
		"Limit" : limit,
		"Search": search,
		"Role": role,
		"UnreadErrors": unreadErrors,
	}

	ac.views.RenderPage(w, r, "admin-user-list", pageData)
}

// DetailUser menampilkan halaman detail read-only untuk seorang pengguna.
func (ac *AdminController) DetailUser(w http.ResponseWriter, r *http.Request) {
	idstr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, "ID pengguna tidak valid", http.StatusBadRequest)
		return
	}

	user, err := models.FindUserByID(ac.env.DB, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal mengambil data pengguna: %v", err), http.StatusInternalServerError)
		return
	}

	profile := map[string]interface{}{
		"ID":      user.ID,
		"Name":    user.Name,
		"Email":   user.Email,
		"Avatar":  user.Avatar,
		"Address": user.Address,
		"Phone":   user.Phone,
		"NIM":     nil,
		"NIP":     nil,
		"NUPTK":   nil,
	}

	// Mahasiswa
	if user.Student != nil {
		profile["NIM"] = user.Student.NIM
	}

	// Dosen
	if user.Lecturer != nil {
		profile["NIP"] = user.Lecturer.NIP
		profile["NUPTK"] = user.Lecturer.NUPTK
	}

	role := user.Roles[0].Name

	var positions []string
	if role == "dosen" {
		positionsDetails, err := models.GetLecturerPositionsByLecturerID(ac.env.DB, user.Lecturer.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Gagal mengambil data posisi dosen: %v", err), http.StatusInternalServerError)
			return
		}

		for _, pos := range positionsDetails {
			if pos.Scopetype == "none" {
				positions = append(positions, pos.PositionName)
			} else {
				positions = append(positions, fmt.Sprintf("%s (%s: %s)", pos.PositionName, pos.Scopetype, pos.ScopeName.String))
			}
		}
	}

	data := map[string]interface{}{
		"User":        user,
		"Role":        role,
		"Positions":   positions,
		"Profile":     profile,
	}

	ac.views.RenderPage(w, r, "admin-user-detail", data)
}

// StreamUserSync menangani SSE untuk Sinkronisasi User (Dummy Logic)
func (ac *AdminController) StreamUserSync(w http.ResponseWriter, r *http.Request) {
    // 1. Setup Header SSE
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming unsupported", 500)
        return
    }

    // Helper kirim JSON ke browser
    sendUpdate := func(progress int, logMsg, status string) {
        data := map[string]interface{}{
            "progress": progress,
            "log":      logMsg,
            "status":   status,
        }
        jsonMsg, _ := json.Marshal(data)
        fmt.Fprintf(w, "data: %s\n\n", jsonMsg)
        flusher.Flush()
    }

    // --- MULAI LOGIKA SINKRONISASI SEMENTARA ---
    
    // Catat log DB: Start
    models.CreateLog(ac.env.DB, "MANUAL", "USER", "RUNNING", "Memulai sinkronisasi user manual.")

    sendUpdate(5, "Menghubungkan ke API Data Center...", "running")
    time.Sleep(1 * time.Second) // Simulasi delay koneksi

    sendUpdate(10, "Koneksi berhasil. Meminta data /users/sync...", "running")
    time.Sleep(500 * time.Millisecond)

    // Simulasi Pagination Loop
    totalPages := 5
    for i := 1; i <= totalPages; i++ {
        // Simulasi fetch API
        offset := (i - 1) * 100
        sendUpdate(10 + (i * 15), fmt.Sprintf("⬇️ Fetching Page %d (Offset: %d)...", i, offset), "running")
        
        time.Sleep(800 * time.Millisecond) // Simulasi download

        // Simulasi Insert DB
        sendUpdate(10 + (i * 15) + 5, fmt.Sprintf("✅ Menyimpan %d user ke database lokal...", 100), "running")
        time.Sleep(200 * time.Millisecond) // Simulasi insert query
    }

    // Catat log DB: Sukses
    models.CreateLog(ac.env.DB, "MANUAL", "USER", "SUCCESS", "Berhasil sinkronisasi 500 user.")

    sendUpdate(100, "✨ Sinkronisasi User Selesai!", "done")
}
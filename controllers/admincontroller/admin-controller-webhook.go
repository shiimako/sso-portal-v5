package admincontroller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sso-portal-v3/models"
)

type WebhookPayload struct {
	Event     string          `json:"event"`     // e.g., "user.updated", "jurusan.created"
	Timestamp int64           `json:"timestamp"` // Waktu kejadian
	Data      json.RawMessage `json:"data"`      // Data dinamis (bisa User, Jurusan, dll)
}

// HandleWebhook menerima push data dari Data Center
func (ac *AdminController) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Gagal membaca body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()


	signature := r.Header.Get("X-Signature")
	if !ac.isValidSignature(bodyBytes, signature) {
		models.CreateLog(ac.env.DB, "WEBHOOK", "SYSTEM", "ERROR", "Invalid Signature detected.")
		http.Error(w, "Unauthorized: Invalid Signature", http.StatusForbidden)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var processErr error
	
	switch payload.Event {
	
	// --- CASE: USER ---
	case "user.created", "user.updated":
		var user models.DCUser
		if err := json.Unmarshal(payload.Data, &user); err != nil {
			processErr = fmt.Errorf("gagal decode user: %v", err)
		} else {
			processErr = models.UpsertFullUser(ac.env.DB, user)
		}

	// --- CASE: JURUSAN ---
	case "jurusan.created", "jurusan.updated":
		var mjr models.DCMajor
		if err := json.Unmarshal(payload.Data, &mjr); err != nil {
			processErr = fmt.Errorf("gagal decode jurusan: %v", err)
		} else {
			processErr = models.UpsertMajor(ac.env.DB, mjr)
		}

	// --- CASE: PRODI ---
	case "prodi.created", "prodi.updated":
		var prd models.DCStudyProgram
		if err := json.Unmarshal(payload.Data, &prd); err != nil {
			processErr = fmt.Errorf("gagal decode prodi: %v", err)
		} else {
			processErr = models.UpsertStudyPrograms(ac.env.DB, prd)
		}

	// --- CASE: ROLE ---
	case "role.created", "role.updated":
		var rl models.DCRole
		if err := json.Unmarshal(payload.Data, &rl); err != nil {
			processErr = fmt.Errorf("gagal decode role: %v", err)
		} else {
			processErr = models.UpsertRole(ac.env.DB, rl)
		}
	
	// --- CASE: JABATAN ---
	case "jabatan.created", "jabatan.updated":
		var pos models.DCPosition
		if err := json.Unmarshal(payload.Data, &pos); err != nil {
			processErr = fmt.Errorf("gagal decode jabatan: %v", err)
		} else {
			processErr = models.UpsertPosition(ac.env.DB, pos)
		}

	default:
		// Event tidak dikenal, ignore saja (200 OK)
		fmt.Printf("Webhook event ignored: %s\n", payload.Event)
	}

	// 6. Logging Hasil
	if processErr != nil {
		fmt.Printf("❌ Webhook Error [%s]: %v\n", payload.Event, processErr)
		models.CreateLog(ac.env.DB, "WEBHOOK", "ALL", "ERROR", fmt.Sprintf("[%s] %v", payload.Event, processErr))
		http.Error(w, "Processing Failed", 500)
		return
	}

	// Sukses
	models.CreateLog(ac.env.DB, "WEBHOOK", "ALL", "SUCCESS", fmt.Sprintf("Event %s berhasil di-sync.", payload.Event))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook Received"))
}

// Helper: Cek apakah signature valid
func (ac *AdminController) isValidSignature(body []byte, signature string) bool {
	if ac.env.WebhookSecret == "" {
		return true // Kalau secret kosong (dev mode), loloskan saja (hati-hati di prod!)
	}
	
	mac := hmac.New(sha256.New, []byte(ac.env.WebhookSecret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSig := hex.EncodeToString(expectedMAC)

	// Compare aman (timing attack safe)
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}
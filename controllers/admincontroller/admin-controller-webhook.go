package admincontroller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sso-portal-v5/models"
	"sso-portal-v5/services"
	"strings"
)

type WebhookPayload struct {
	Event     string          `json:"event"`     
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`     
}

// HandleWebhook menerima push data 
func (ac *AdminController) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("HTTP ERROR path=%s, err=%v", r.URL.Path, "Invalid Method, Expected POST")
		http.Error(w, "Invalid Method, Expected POST", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to Read Body", http.StatusBadRequest)
		log.Printf("HTTP ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}
	defer r.Body.Close()

	senderIP := ac.getClientIP(r)
	signature := r.Header.Get("X-Signature")
	
	if !ac.isValidSignature(bodyBytes, signature) {
		log.Printf("SECURITY ALERT: Webhook Invalid Signature from IP=%s", senderIP)
		
		user, err := models.FindUserByEmail(ac.env.DB, ac.env.AdminEmail)
		if err == nil {
			go services.SendPushNotification(
				ac.env,
				user.ID,
				"Portal Security",
				"Unauthorized Webhook Signature Detected!",
				ac.env.BaseURL,
			)
		}

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

	case "notification.push":
		var notifData struct {
			Email   string `json:"email"`
			AppSlug string `json:"app_slug"`
			Count   int    `json:"count"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(payload.Data, &notifData); err != nil {
			processErr = fmt.Errorf("gagal decode notif: %v", err)
		} else {
			user, err := models.FindUserByEmail(ac.env.DB, notifData.Email)
			if err != nil {
				processErr = fmt.Errorf("user email tidak ditemukan: %s", notifData.Email)
			} else {
				app, err := models.FindApplicationBySlug(ac.env.DB, notifData.AppSlug)
				if err != nil {
					processErr = fmt.Errorf("app slug not found: %s", notifData.AppSlug)
				} else {
					// Simpan ke DB
					processErr = models.InsertNotification(ac.env.DB, user.ID, app.ID, notifData.Message)

					// Trigger Service Push (Realtime)
					go services.SendPushNotification(
						ac.env,
						user.ID,
						app.Name,
						notifData.Message,
						fmt.Sprintf("%s/redirect?app=%s", ac.env.BaseURL, notifData.AppSlug),
					)
				}
			}
		}

	default:
		log.Printf("Webhook Ignored (Standalone Mode): %s", payload.Event)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Event Ignored"))
		return
	}

	if processErr != nil {
		fmt.Printf("❌ Webhook Error [%s]: %v\n", payload.Event, processErr)
		http.Error(w, "Processing Failed", 500)
		return
	}

	// Sukses
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification Received"))
}

// Helper: Cek apakah signature valid
func (ac *AdminController) isValidSignature(body []byte, signature string) bool {
	if ac.env.WebhookSecret == "" {
		return true
	}

	mac := hmac.New(sha256.New, []byte(ac.env.WebhookSecret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSig := hex.EncodeToString(expectedMAC)

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// Helper: Ambil IP Client
func (ac *AdminController) getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	realIP := r.Header.Get("X-Real-Ip")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
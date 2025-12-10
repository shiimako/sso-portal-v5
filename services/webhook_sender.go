package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sso-portal-v3/config"
	"sso-portal-v3/models"
	"time"
)

// SendUserUpdateWebhook mengirim notifikasi ke Data Center
func SendUserUpdateWebhook(env *config.Env, user *models.FullUser, updatedData map[string]interface{}) {

	targetURL := fmt.Sprintf("%s/webhook/listener", env.DataCenterURL)

	payload := map[string]interface{}{
		"event":     "user.profile_updated_from_portal",
		"timestamp": time.Now().Unix(),
		"source":    "sso_portal",
		"user_id":   user.ID,
		"email":     user.Email,
		"changes":   updatedData,
	}

	jsonBody, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(env.WebhookSecret))
	mac.Write(jsonBody)
	signature := hex.EncodeToString(mac.Sum(nil))

	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			fmt.Printf("❌ Gagal membuat request webhook: %v\n", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Signature", signature)
		req.Header.Set("X-API-KEY", env.DataCenterKey)
		req.Header.Set("X-Client-ID", "sso-portal-v3")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Gagal mengirim webhook ke Data Center: %v\n", err)
			return
		}
		defer resp.Body.Close()

		fmt.Printf("🚀 Webhook Profile Update dikirim! Status: %d\n", resp.StatusCode)
	}()
}

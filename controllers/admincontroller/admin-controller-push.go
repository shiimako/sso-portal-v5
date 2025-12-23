package admincontroller

import (
	"encoding/json"
	"log"
	"net/http"
	"sso-portal-v5/models"
)

type SubscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func (ac *AdminController) SubscribePush(w http.ResponseWriter, r *http.Request) {
	userLogin := r.Context().Value("UserLogin").(*models.FullUser)

	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ac.RenderError(w, r, http.StatusBadRequest, "Format Data Tidak Valid")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	query := `
		INSERT INTO user_push_subscriptions (user_id, endpoint, p256dh, auth, created_at)
		VALUES (?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE 
			user_id = VALUES(user_id),
			updated_at = NOW()
	`
	
	_, err := ac.env.DB.Exec(query, userLogin.ID, req.Endpoint, req.Keys.P256dh, req.Keys.Auth)
	if err != nil {
		ac.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	w.Write([]byte("Subscribed successfully"))
}
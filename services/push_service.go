package services

import (
	"encoding/json"
	"log"
	"sso-portal-v5/config"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type PushSub struct {
	Endpoint string `db:"endpoint"`
	P256dh   string `db:"p256dh"`
	Auth     string `db:"auth"`
}

func SendPushNotification(env *config.Env, userID int, title, message, url string) {
	var subs []PushSub
	err := env.DB.Select(&subs, "SELECT endpoint, p256dh, auth FROM user_push_subscriptions WHERE user_id = ?", userID)
	if err != nil {
		log.Println("ERROR [Send Push]: ", err)
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"title": title,
		"body":  message,
		"url":   url,
		"icon":  "/public/static/Logo-PNC.png",
	})

	for _, sub := range subs {
		s := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}

		resp, err := webpush.SendNotification(payload, s, &webpush.Options{
			Subscriber:      env.VapidSubject,
			VAPIDPublicKey:  env.VapidPublicKey,
			VAPIDPrivateKey: env.VapidPrivateKey,
			TTL:             30,
		})

		if err != nil {
			log.Println("Gagal kirim push:", err)
		} else {
			resp.Body.Close()
		}
	}
}
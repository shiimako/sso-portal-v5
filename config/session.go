// file: config/session.go

package config

import (
	"log"
	"os"

	"github.com/gorilla/sessions"
)

// InitSessionStore menginisialisasi cookie store untuk session.
func InitSessionStore() *sessions.CookieStore {
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("FATAL: SESSION_KEY tidak ditemukan di environment variables.")
	}

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		// Secure: true, // <-- Aktifkan ini saat sudah menggunakan HTTPS (production)
	}

	return store
}
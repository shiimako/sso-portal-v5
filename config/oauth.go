// file: config/oauth.go

package config

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// InitGoogleOAuthConfig menginisialisasi konfigurasi OAuth2 untuk Google.
func InitGoogleOAuthConfig(baseURL string) *oauth2.Config {
	googleOAuthConfig := &oauth2.Config{
		// Ambil kredensial dari environment variables
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),

		// Redirect URL harus sama persis dengan yang didaftarkan di Google Cloud Console
		RedirectURL:  baseURL + "/auth/google/callback",

		// Endpoint Google sudah disediakan oleh package
		Endpoint:     google.Endpoint,

		// Scopes menentukan data apa yang kita minta izinnya dari pengguna
		Scopes:       []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},

	}

	return googleOAuthConfig
}
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

		RedirectURL:  baseURL + "/auth/google/callback",

		Endpoint:     google.Endpoint,
		Scopes:       []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},

	}

	return googleOAuthConfig
}
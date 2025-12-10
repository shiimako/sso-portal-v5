package config

import (
	"html/template"
	"os"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type Env struct {
	DB        *sqlx.DB
	Store     *sessions.CookieStore
	Templates map[string]*template.Template
	SessionName string
	BaseURL   string
	GoogleOAuthConfig *oauth2.Config
	AdminEmail string

	// Data Center Config
	DataCenterURL string
	DataCenterKey string
	WebhookSecret string

}

func NewEnv(db *sqlx.DB, store *sessions.CookieStore, templates map[string]*template.Template) *Env {
	dcURL := os.Getenv("DATA_CENTER_URL")
	if dcURL == "" {
		dcURL = "http://localhost:8000/api/v1" 
	}

	return &Env{
		DB:                db,
		Store:             store,
		Templates:         templates,
		SessionName:       os.Getenv("SESSION_NAME"),
		BaseURL:           os.Getenv("APP_BASE_URL"),
		GoogleOAuthConfig: InitGoogleOAuthConfig(os.Getenv("APP_BASE_URL")),
		AdminEmail: os.Getenv("ADMIN_EMAIL_OVERRIDE"),
		
		// Load Data Center Config
		DataCenterURL: dcURL,
		DataCenterKey: os.Getenv("DATA_CENTER_KEY"),
		WebhookSecret: os.Getenv("WEBHOOK_SECRET"),
	}
}


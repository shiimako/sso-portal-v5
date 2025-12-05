package config

import (
	"html/template"

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
}


// file: handlers/env.go

package handlers

import (
	"html/template"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

// Env adalah struct yang berfungsi sebagai "kotak perkakas".
// Struct ini akan menampung semua dependensi (koneksi db, session store, dll)
// yang dibutuhkan oleh para handler/controller.
type Env struct {
	DB        *sqlx.DB
	Store     *sessions.CookieStore
	Templates map[string]*template.Template
	SessionName string
	BaseURL   string
	GoogleOAuthConfig *oauth2.Config
}


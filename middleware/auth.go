package middleware

import (
	"context"
	"log"
	"net/http"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"
)

func GlobalAuthMiddleware(env *handlers.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, _ := env.Store.Get(r, env.SessionName)

			auth, ok := session.Values["authenticated"].(bool)
			if !ok || !auth {
				session.AddFlash("Anda harus login terlebih dahulu.")
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			rawID, ok := session.Values["user_id"]
			if !ok {
				http.Error(w, "User ID tidak valid", http.StatusUnauthorized)
				return
			}

			userID := rawID.(int)
			user, err := models.FindUserByID(env.DB, userID)
			if err != nil {
				http.Error(w, "Gagal mengambil data user", http.StatusInternalServerError)
				return
			}
			role := user.Roles[0].Name

			ctx := context.WithValue(r.Context(), "UserLogin", user)
			ctx = context.WithValue(ctx, "ActiveRole", role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(env *handlers.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Ambil data user dari context (hasil GlobalAuthMiddleware)
			userLogin, ok := r.Context().Value("UserLogin").(*models.FullUser)
			if !ok || userLogin == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				log.Println("userLogin : ", userLogin)
				return
			}

			role, ok := r.Context().Value("ActiveRole").(string)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				log.Println("role : ", role)
				return
			}

			// 2. Cek role
			if role != "admin" {
				http.Error(w, "Akses ditolak. Halaman ini hanya untuk admin.", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


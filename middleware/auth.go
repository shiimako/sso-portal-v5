package middleware

import (
	"context"
	"log"
	"net/http"
	"sso-portal-v5/config"
	"sso-portal-v5/models"
	"sso-portal-v5/views"
)

func GlobalAuthMiddleware(env *config.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, _ := env.Store.Get(r, env.SessionName)

			auth, ok := session.Values["authenticated"].(bool)
			if !ok || !auth {
				session.AddFlash("Anda harus login terlebih dahulu.")
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			rawID, ok := session.Values["user_id"]
			if !ok {
				session.AddFlash("User ID Tidak Valid")
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			userID := rawID.(int)
			user, err := models.FindUserByID(env.DB, userID)
			if err != nil {
				session.AddFlash("Gagal Mengambil ID User, hubungi Administrator")
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
				return
			}

			if user.Status != "aktif" {
				session.Values["authenticated"] = false
                delete(session.Values, "user_id")
                session.Options.MaxAge = -1
				
				session.AddFlash("Akun Anda Tidak Aktif. Silahkan Hubungi Administrator.")
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			role := user.Roles[0].Name

			ctx := context.WithValue(r.Context(), "UserLogin", user)
			ctx = context.WithValue(ctx, "ActiveRole", role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(env *config.Env, v *views.Views) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

            roleVal := r.Context().Value("ActiveRole")
            role, ok := roleVal.(string)

            if !ok || role != "admin" {
                w.WriteHeader(http.StatusForbidden)
                data := map[string]interface{}{
                    "Code":    http.StatusForbidden,
                    "Message": "Akses ditolak. Anda tidak memiliki izin Administrator untuk mengakses halaman ini.",
                }
                v.RenderPage(w, r, "error", data) 
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

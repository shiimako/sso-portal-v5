// file: middleware/auth.go

package middleware

import (
	"net/http"
	"sso-portal-v3/handlers" // Sesuaikan path jika perlu
)

// AdminMiddleware memastikan hanya pengguna dengan peran aktif 'admin' yang bisa lewat.
func AdminMiddleware(env *handlers.Env) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := env.Store.Get(r, env.SessionName)

			// 1. Cek apakah pengguna sudah login
			if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
				session.AddFlash("Anda harus login untuk mengakses halaman ini.")
				session.Save(r, w)
				http.Redirect(w, r, "/", http.StatusFound)
				return // Hentikan proses
			}

			// 2. Cek apakah peran aktif adalah 'admin'
			activeRole, ok := session.Values["active_role"].(string)
			if !ok || activeRole != "admin" {
				// Jika bukan admin, tolak dengan error 403 Forbidden (Akses Ditolak)
				http.Error(w, "Akses Ditolak: Halaman ini hanya untuk administrator.", http.StatusForbidden)
				return // Hentikan proses
			}

			// Jika lolos semua pemeriksaan, lanjutkan ke handler tujuan
			next.ServeHTTP(w, r)
		})
	}
}
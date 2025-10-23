package redirectcontroller

import (
	"fmt"
	"net/http"
	"os"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RedirectController struct {
	env *handlers.Env
}

func NewRedirectController(env *handlers.Env) *RedirectController {
	return &RedirectController{env: env}
}

// Claims adalah struktur data (Payload) di dalam JWT.
type Claims struct {
	Name       string                   `json:"name"`
	Email      string                   `json:"email"`
	Avatar     string                   `json:"avatar"`
	BaseRole   string                   `json:"base_role"`
	Attributes []map[string]interface{} `json:"attributes"`
	Profile    map[string]string        `json:"profile"`
	jwt.RegisteredClaims
}

// RedirectToApp membuat JWT dan mengarahkan pengguna ke aplikasi tujuan.
func (rc *RedirectController) RedirectToApp(w http.ResponseWriter, r *http.Request) {
	session, _ := rc.env.Store.Get(r, rc.env.SessionName)

	// 1. Validasi: Pastikan pengguna sudah login & punya peran aktif
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Tidak terautentikasi", http.StatusUnauthorized)
		return
	}
	activeRole, ok := session.Values["active_role"].(string)
	if !ok || activeRole == "" {
		http.Redirect(w, r, "/select-role", http.StatusFound)
		return
	}

	// 2. Ambil semua data pengguna dari session
	userID, ok := session.Values["user_id"].(int)
    if !ok {
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }
	avatarURL := session.Values["avatar"].(string)

	var attributes []map[string]interface{}

	if activeRole == "dosen" {
		if val, ok := session.Values["attributes"].([]map[string]interface{}); ok {
			attributes = val
		} 
	}

		user, err := models.FindUserByID(rc.env.DB, fmt.Sprintf("%d", userID))
		if err != nil {
			http.Error(w, "Gagal mengambil detail pengguna", http.StatusInternalServerError)
			return
		}

		// 3. Ambil slug aplikasi tujuan dari URL
		appSlug := r.URL.Query().Get("app")
		if appSlug == "" {
			http.Error(w, "Aplikasi tujuan tidak spesifik", http.StatusBadRequest)
			return
		}

		// 4. Ambil detail aplikasi dari database
		app, err := models.FindApplicationBySlug(rc.env.DB, appSlug)
		if err != nil {
			http.Error(w, "Aplikasi tidak ditemukan atau tidak terdaftar", http.StatusNotFound)
			return
		}

		// 5. Ambil data profil spesifik (NIM/NIP/NIDN)
		profileData := make(map[string]string)
		if user.NIM.Valid {
			profileData["nim"] = user.NIM.String
		}
		if user.NIP.Valid {
			profileData["nip"] = user.NIP.String
		}
		if user.NIDN.Valid {
			profileData["nidn"] = user.NIDN.String
		}
		if user.Address.Valid {
			profileData["address"] = user.Address.String
		}
		if user.Phone.Valid {
			profileData["phone"] = user.Phone.String
		}

		// 6. Buat "Halaman Identitas" (Payload/Claims) untuk paspor
		expirationTime := time.Now().Add(10 * time.Second) // Paspor berlaku 1 jam
		claims := &Claims{
			Name:       user.Name,
			Email:      user.Email,
			Avatar:     avatarURL,
			BaseRole:   activeRole,
			Attributes: attributes,
			Profile:    profileData,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
				Subject:   fmt.Sprintf("%d", userID),
				Issuer:    "SSO-PNC", // Nama penerbit paspor
			},
		}

		// 7. Siapkan "Sampul Paspor" dan masukkan "Halaman Identitas"
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// 8. Tandatangani paspor dengan kunci rahasia (membuat hologram)
		jwtSecret := os.Getenv("JWT_SECRET_KEY")
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
			return
		}

		// 9. Arahkan pengguna ke aplikasi tujuan sambil membawa paspor (token)
		finalURL := fmt.Sprintf("%s?token=%s", app.TargetURL, tokenString)
		http.Redirect(w, r, finalURL, http.StatusTemporaryRedirect)
	}

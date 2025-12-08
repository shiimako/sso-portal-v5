package authcontroller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sso-portal-v3/config"
	"sso-portal-v3/models"
	"strings"
)

type AuthController struct {
	env *config.Env
}

func NewAuthController(env *config.Env) *AuthController {
	return &AuthController{env: env}
}

// ShowLoginPage menampilkan halaman login.
func (ac *AuthController) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	flashes := session.Flashes()
	session.Save(r, w)
	data := map[string]interface{}{
		"FlashMessages": flashes,
	}

	err := ac.env.Templates["login"].ExecuteTemplate(w, "login.html", data)
	if err != nil {
		log.Printf("Gagal render template login: %v", err)
		http.Error(w, "Terjadi kesalahan internal", http.StatusInternalServerError)
	}
}

// LoginWithGoogle menginisiasi proses OAuth2 dengan Google.
func (ac *AuthController) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("ERROR: Gagal membuat state token acak: %v", err)
		http.Error(w, "Gagal membuat state token", http.StatusBadRequest)
		return
	}
	state := base64.URLEncoding.EncodeToString(b)

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Gagal menyimpan session", http.StatusInternalServerError)
		return
	}

	url := ac.env.GoogleOAuthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback menangani callback dari Google setelah user mengotorisasi aplikasi.
func (ac *AuthController) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	sessionState, ok := session.Values["state"].(string)

	if !ok {
		http.Error(w, "State tidak ditemukan di session", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != sessionState {
		http.Error(w, "State tidak cocok", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")

	token, err := ac.env.GoogleOAuthConfig.Exchange(context.Background(), code)

	if err != nil {
		log.Printf("ERROR: Gagal menukar kode dengan token: %v", err)
		http.Error(w, "Gagal menukar kode dengan token", http.StatusInternalServerError)
		return
	}

	client := ac.env.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Gagal mendapatkan user info dari Google: %v", err)
		http.Error(w, "Gagal mendapatkan informasi user dari Google", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userProfile struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Picture       string `json:"picture"`
	}

	err = json.NewDecoder(resp.Body).Decode(&userProfile)
	if err != nil {
		log.Printf("ERROR: Gagal decode user info JSON: %v", err)
		http.Error(w, "Gagal memproses informasi user dari Google", http.StatusInternalServerError)
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL_OVERRIDE")
	if !userProfile.VerifiedEmail || (!strings.HasSuffix(userProfile.Email, "@pnc.ac.id") && userProfile.Email != adminEmail) {
		session.AddFlash("Hanya email dengan domain @pnc.ac.id yang diizinkan.")
		session.Save(r, w)
		log.Println("Akses ditolak untuk email:", userProfile.Email)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user, err := models.FindUserByEmail(ac.env.DB, userProfile.Email)
	if err != nil {
		http.Error(w, "Gagal mengambil detail user", http.StatusInternalServerError)
		log.Println("ERROR : ", err)
		return
	}

	if user == nil || user.ID == 0 {
		session.AddFlash("Email Anda belum terdaftar di sistem.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		log.Println("Akses ditolak, user tidak ditemukan:", userProfile.Email)
		return
	}

	if user.Status != "aktif" {
		session.AddFlash("Akun Anda tidak aktif. Silakan hubungi administrator.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		log.Println("Akses ditolak, user tidak aktif:", userProfile.Email)
		return
	}

	avatarEmpty := !user.Avatar.Valid || strings.TrimSpace(user.Avatar.String) == ""

	if avatarEmpty && userProfile.Picture != "" {
		log.Printf("INFO: Avatar untuk user %d kosong, mengisi dari Google...", user.ID)
		err = models.UpdateUserAvatar(ac.env.DB, user.ID, userProfile.Picture)
		if err != nil {
			log.Printf("WARNING: Gagal update avatar untuk user %d: %v", user.ID, err)
		}
	}

	session.Values["authenticated"] = true
	session.Values["user_id"] = user.ID

	if user.Avatar.Valid {
		session.Values["avatar"] = user.Avatar.String
	} else {
		session.Values["avatar"] = userProfile.Picture
	}

	delete(session.Values, "state")

	if len(user.Roles) == 0 {
		session.AddFlash("Akun Anda tidak memiliki peran. Hubungi administrator.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	session.Save(r, w)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// Logout menghapus session pengguna
func (ac *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	// Cara terbaik untuk menghapus session adalah dengan membuatnya kedaluwarsa.
	// Mengatur MaxAge ke -1 akan memberitahu browser untuk segera menghapus cookie.
	session.Options.MaxAge = -1

	// Simpan perubahan pada session
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, "Gagal untuk logout", http.StatusInternalServerError)
		return
	}

	// Arahkan pengguna kembali ke halaman login
	http.Redirect(w, r, "/", http.StatusFound)
}

package authcontroller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sso-portal-v5/config"
	"sso-portal-v5/models"
	"sso-portal-v5/views"
	"strings"
)

type AuthController struct {
	env *config.Env
	views *views.Views
}

func NewAuthController(env *config.Env, v *views.Views) *AuthController {
	return &AuthController{env: env, views: v}
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

	ac.env.Templates["login"].ExecuteTemplate(w, "login.html", data)
}

// LoginWithGoogle menginisiasi proses OAuth2 dengan Google.
func (ac *AuthController) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}
	state := base64.URLEncoding.EncodeToString(b)

	session, _ := ac.env.Store.Get(r, ac.env.SessionName)
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
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
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		http.Error(w, "Gagal menukar kode dengan token", http.StatusInternalServerError)
		return
	}

	client := ac.env.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
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
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		http.Error(w, "Gagal memproses informasi user dari Google", http.StatusInternalServerError)
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL_OVERRIDE")
	if !userProfile.VerifiedEmail || (!strings.HasSuffix(userProfile.Email, "@pnc.ac.id") && userProfile.Email != adminEmail) {
		session.AddFlash("Hanya email dengan domain @pnc.ac.id yang diizinkan.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user, err := models.FindUserByEmail(ac.env.DB, userProfile.Email)
	if err != nil {
		http.Error(w, "Gagal mengambil detail user", http.StatusInternalServerError)
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
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

	session.Options.MaxAge = -1

	err := session.Save(r, w)
	if err != nil {
		http.Error(w, "Gagal untuk logout", http.StatusInternalServerError)
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	// Arahkan pengguna kembali ke halaman login
	http.Redirect(w, r, "/", http.StatusFound)
}

func (ac *AuthController) RenderError(w http.ResponseWriter, r *http.Request, code int, message string) {
    w.WriteHeader(code)
    
    data := map[string]interface{}{
        "Code":    code,
        "Message": message,
    }
    
    ac.views.RenderPage(w, r, "error", data)
}

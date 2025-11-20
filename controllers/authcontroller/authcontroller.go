package authcontroller

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"
	"strings"
)

type UserRole struct {
	Name  string
	Type  string
	Scope sql.NullString
}

type AuthController struct {
	env *handlers.Env
}

func NewAuthController(env *handlers.Env) *AuthController {
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

	err := ac.env.Templates.ExecuteTemplate(w, "login.html", data)
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
	if !userProfile.VerifiedEmail || !strings.HasSuffix(userProfile.Email, "@pnc.ac.id") && userProfile.Email != adminEmail {
		session.AddFlash("Hanya email dengan domain @pnc.ac.id yang diizinkan.")
		session.Save(r, w)
		log.Println("Akses ditolak untuk email:", userProfile.Email)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	var userID int
	var userStatus string
	var userName string
	var allRoles []UserRole

	query := `SELECT u.id, u.name, u.status, r.name, r.type, ur.scope FROM users u 
	           JOIN user_roles ur ON u.id = ur.user_id 
	           JOIN roles r ON ur.role_id = r.id 
	           WHERE LOWER(TRIM(u.email)) = LOWER(?)`

	rows, err := ac.env.DB.Query(query, userProfile.Email)
	if err != nil {
		http.Error(w, "Gagal query ke database", http.StatusInternalServerError)
		log.Println("ERROR: Gagal query ke database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var role UserRole
		err = rows.Scan(&userID, &userName, &userStatus, &role.Name, &role.Type, &role.Scope)
		if err != nil {
			http.Error(w, "Gagal memproses data user", http.StatusInternalServerError)
			log.Println("ERROR: Gagal scan data user:", err)
			return
		}
		allRoles = append(allRoles, role)
	}

	if userID == 0 {
		session.AddFlash("Email Anda belum terdaftar di sistem.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		log.Println("Akses ditolak, user tidak ditemukan:", userProfile.Email)
		return
	}

	if userStatus != "aktif" {
		session.AddFlash("Akun Anda tidak aktif. Silakan hubungi administrator.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
		log.Println("Akses ditolak, user tidak aktif:", userProfile.Email)
		return
	}

	user, err := models.FindUserByID(ac.env.DB, fmt.Sprintf("%d", userID))
	if err != nil {
		// Handle error (meskipun kecil kemungkinannya karena kita baru cek email)
		http.Error(w, "Gagal mengambil detail user", http.StatusInternalServerError)
		return
	}

	if !user.Avatar.Valid && userProfile.Picture != "" {
		log.Printf("INFO: Avatar untuk user %d kosong, mengisi dari Google...", user.ID)
		err = models.UpdateUserAvatar(ac.env.DB, user.ID, userProfile.Picture)
		if err != nil {
			// Bukan error fatal, cukup log saja
			log.Printf("WARNING: Gagal update avatar untuk user %d: %v", user.ID, err)
		}
	}

	var baseRoles []string

	for _, role := range allRoles {
		if role.Type == "base" {
			baseRoles = append(baseRoles, role.Name)
		} 
	}

	session.Values["authenticated"] = true
	session.Values["user_id"] = userID
	session.Values["avatar"] = userProfile.Picture
	delete(session.Values, "state")

	if len(baseRoles) == 0 {
		session.AddFlash("Akun Anda tidak memiliki peran. Hubungi administrator.")
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		session.Values["active_role"] = baseRoles[0]
		session.Save(r, w)
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

// SelectRolePage menampilkan halaman pemilihan peran jika user memiliki banyak peran.
func (ac *AuthController) SelectRolePage(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	availableRoles, ok := session.Values["available_roles"].([]string)
	if !ok || len(availableRoles) == 0 {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"Roles": availableRoles,
	}

	err := ac.env.Templates.ExecuteTemplate(w, "select-role.html", data)
	if err != nil {
		log.Printf("Gagal render template select_role: %v", err)
		http.Error(w, "Terjadi kesalahan internal", http.StatusInternalServerError)
	}
}

// SelectRoleHandler menangani pemilihan peran oleh user.
func (ac *AuthController) SelectRoleHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := ac.env.Store.Get(r, ac.env.SessionName)

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	chosenRole := r.URL.Query().Get("role")
	if chosenRole == "" {
		http.Error(w, "Peran tidak dipilih", http.StatusBadRequest)
		return
	}

	avaiableRoles, ok := session.Values["available_roles"].([]string)
	isAllowed := false
	for _, role := range avaiableRoles {
		if role == chosenRole {
			isAllowed = true
			break
		}
	}
	if !ok || !isAllowed {
		http.Error(w, "Anda tidak memiliki hak akses untuk peran ini", http.StatusForbidden)
		return
	}

	session.Values["active_role"] = chosenRole
	delete(session.Values, "active_roles")
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

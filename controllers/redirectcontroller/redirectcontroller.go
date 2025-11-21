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
	Name    string            `json:"name"`
	Email   string            `json:"email"`
	Avatar  string            `json:"avatar"`
	Role    string            `json:"role"`
	Profile map[string]string `json:"profile"`
	jwt.RegisteredClaims
}

// RedirectToApp membuat JWT dan mengarahkan pengguna ke aplikasi tujuan.
func (rc *RedirectController) RedirectToApp(w http.ResponseWriter, r *http.Request) {
	session, _ := rc.env.Store.Get(r, rc.env.SessionName)

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Tidak terautentikasi", http.StatusUnauthorized)
		return
	}
	activeRole := session.Values["active_role"].(string)

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	user, err := models.FindUserByID(rc.env.DB, userID)
	if err != nil {
		http.Error(w, "Gagal mengambil detail pengguna", http.StatusInternalServerError)
		return
	}

	appSlug := r.URL.Query().Get("app")
	if appSlug == "" {
		http.Error(w, "Aplikasi tujuan tidak spesifik", http.StatusBadRequest)
		return
	}

	app, err := models.FindApplicationBySlug(rc.env.DB, appSlug)
	if err != nil {
		http.Error(w, "Aplikasi tidak ditemukan atau tidak terdaftar", http.StatusNotFound)
		return
	}

	profileData := make(map[string]string)
	if user.Student.ID != 0 {
		profileData["student_id"] = fmt.Sprintf("%d", user.Student.ID)
	}
	if user.Lecturer.ID != 0 {
		profileData["lecturer_id"] = fmt.Sprintf("%d", user.Lecturer.ID)
	}

	expirationTime := time.Now().Add(10 * time.Second)
	claims := &Claims{
		Name:    user.Name,
		Email:   user.Email,
		Avatar:  user.Avatar.String,
		Role:    activeRole,
		Profile: profileData,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    "SSO-PNC",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}
	finalURL := fmt.Sprintf("%s?token=%s", app.TargetURL, tokenString)
	http.Redirect(w, r, finalURL, http.StatusTemporaryRedirect)
}

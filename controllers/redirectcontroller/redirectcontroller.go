package redirectcontroller

import (
	"fmt"
	"net/http"
	"os"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"
	"sso-portal-v3/views"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RedirectController struct {
	env   *handlers.Env
	views *views.Views
}

func NewRedirectController(env *handlers.Env, v *views.Views) *RedirectController {
	return &RedirectController{env: env, views: v}
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

	user := r.Context().Value("UserLogin").(*models.FullUser)
	role := user.Roles[0].Name

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
	if user.Student != nil && user.Student.ID != 0 {
		profileData["student_id"] = fmt.Sprintf("%d", user.Student.ID)
	}

	if user.Lecturer != nil && user.Lecturer.ID != 0 {
		profileData["lecturer_id"] = fmt.Sprintf("%d", user.Lecturer.ID)
	}

	expirationTime := time.Now().Add(2 * time.Minute)
	claims := &Claims{
		Name:    user.Name,
		Email:   user.Email,
		Avatar:  fmt.Sprintf("%s/avatar/%d", os.Getenv("APP_BASE_URL"), user.ID),
		Role:    role,
		Profile: profileData,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Subject:   fmt.Sprintf("%d", user.ID),
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

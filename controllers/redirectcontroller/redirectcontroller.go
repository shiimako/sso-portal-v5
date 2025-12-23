package redirectcontroller

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sso-portal-v5/config"
	"sso-portal-v5/models"
	"sso-portal-v5/views"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RedirectController struct {
	env   *config.Env
	views *views.Views
}

func NewRedirectController(env *config.Env, v *views.Views) *RedirectController {
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
		rc.RenderError(w, r, http.StatusBadRequest, "Slug tidak Valid")
		return
	}

	app, err := models.FindApplicationBySlug(rc.env.DB, appSlug)
	if err != nil {
		rc.RenderError(w, r, http.StatusNotFound, "Aplikasi Tidak Ditemukan.")
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
			Issuer:   config.Issuer,
			Audience:  jwt.ClaimStrings{appSlug},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(config.PrivateKey)
	if err != nil {
		rc.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}
	finalURL := fmt.Sprintf("%s?token=%s", app.TargetURL, url.QueryEscape(tokenString))

	go models.ClearNotification(rc.env.DB, user.ID, app.ID)

	http.Redirect(w, r, finalURL, http.StatusTemporaryRedirect)
}

func (ac *RedirectController) RenderError(w http.ResponseWriter, r *http.Request, code int, message string) {
    w.WriteHeader(code)
    
    data := map[string]interface{}{
        "Code":    code,
        "Message": message,
    }
    
    ac.views.RenderPage(w, r, "error", data)
}

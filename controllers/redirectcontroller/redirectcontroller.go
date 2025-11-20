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
	user, err := models.FindUserByID(rc.env.DB, fmt.Sprintf("%d", userID))
	if err != nil {
		http.Error(w, "Gagal mengambil detail pengguna", http.StatusInternalServerError)
		return
	}

	allroles, err := models.GetUserRolesAndAttributes(rc.env.DB, userID)
	if err != nil {
		http.Error(w, "Gagal mengambil peran pengguna", http.StatusInternalServerError)
		return
	}

	var attributes []map[string]interface{}
	for _, role := range allroles {
		if role.Type == "attribute" {
			attr := map[string]interface{}{"role": role.Name}
			if role.Scope.Valid {
				attr["scope"] = role.Scope.String
			}
			attributes = append(attributes, attr)
		}
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
	if user.Student_ID.Valid {
		profileData["student_id"] = fmt.Sprintf("%d", user.Student_ID.Int64)
	}
	if user.NIM.Valid {
		profileData["nim"] = user.NIM.String
	}
	if user.Lecturer_ID.Valid {
		profileData["lecturer_id"] = fmt.Sprintf("%d", user.Lecturer_ID.Int64)
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

	expirationTime := time.Now().Add(10 * time.Second)
	claims := &Claims{
		Name:       user.Name,
		Email:      user.Email,
		Avatar:     user.Avatar.String,
		BaseRole:   activeRole,
		Attributes: attributes,
		Profile:    profileData,
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

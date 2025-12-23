// file: controllers/usercontroller/usercontroller.go
package usercontroller

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sso-portal-v5/config"
	"sso-portal-v5/models"
	"sso-portal-v5/services"
	"sso-portal-v5/views"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type UserController struct {
	env   *config.Env
	views *views.Views
}

func NewUserController(env *config.Env, v *views.Views) *UserController {
	return &UserController{env: env, views: v}
}

// ShowProfileForm menampilkan halaman edit profil pengguna
func (uc *UserController) ShowProfileForm(w http.ResponseWriter, r *http.Request) {

	 session, _ := uc.env.Store.Get(r, uc.env.SessionName)
    
    flashes := session.Flashes()
    session.Save(r, w) 

    var flashMsg string
    if len(flashes) > 0 {
        flashMsg = flashes[0].(string)
    }

	user := r.Context().Value("UserLogin").(*models.FullUser)

	profile := map[string]interface{}{
		"ID":      user.ID,
		"Name":    user.Name,
		"Email":   user.Email,
		"Avatar":  user.Avatar,
		"Address": user.Address,
		"Phone":   user.Phone,
		"NIM":     nil,
		"NIP":     nil,
		"NUPTK":   nil,
	}

	// Mahasiswa
	if user.Student != nil {
		profile["NIM"] = user.Student.NIM
	}

	// Dosen
	if user.Lecturer != nil {
		profile["NIP"] = user.Lecturer.NIP
		profile["NUPTK"] = user.Lecturer.NUPTK
	}

	data := map[string]interface{}{
		"Profile": profile,
		"Flash": flashMsg,
	}

	uc.views.RenderPage(w, r, "edit-profile", data)
}

// HandleProfileUpdate memproses data form edit profil
func (uc *UserController) HandleProfileUpdate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("UserLogin").(*models.FullUser)
	userID := user.ID

	r.ParseMultipartForm(10 << 20)

	address := r.FormValue("address")
	phone := r.FormValue("phone")
	avatarCropped := r.FormValue("avatar-cropped")

	var avatarPath string = ""
	var saveDir = filepath.Join("public", "uploads", "avatars")

	os.MkdirAll(saveDir, os.ModePerm)

	// ---------------------------------------------------------
	// SKENARIO A: Ada data crop (Base64) -> PRIORITAS UTAMA
	// ---------------------------------------------------------
	if avatarCropped != "" && strings.Contains(avatarCropped, "base64,") {
		parts := strings.Split(avatarCropped, ",")
		if len(parts) == 2 {
			dec, err := base64.StdEncoding.DecodeString(parts[1])
			if err == nil {
				fileName := fmt.Sprintf("user-%d-avatar-%d.jpg", userID, time.Now().UnixNano())
				savePath := filepath.Join(saveDir, fileName)

				f, err := os.Create(savePath)
				if err == nil {
					defer f.Close()
					if _, err := f.Write(dec); err == nil {
						avatarPath = "/uploads/avatars/" + fileName
						log.Printf("INFO: User %d update avatar via CROP (Base64)", userID)
					}
				}

			} else {
				log.Printf("ERROR: Gagal decode base64 avatar: %v", err)
			}
		}
	}

	// ---------------------------------------------------------
	// SKENARIO B: Tidak ada crop, tapi ada file upload biasa (Fallback)
	// ---------------------------------------------------------
	if avatarPath == "" {
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()

			ext := filepath.Ext(header.Filename)
			fileName := fmt.Sprintf("user-%d-avatar-%d%s", userID, time.Now().UnixNano(), ext)
			savePath := filepath.Join(saveDir, fileName)

			dst, err := os.Create(savePath)
			if err == nil {
				defer dst.Close()
				if _, err := io.Copy(dst, file); err == nil {
					avatarPath = "/uploads/avatars/" + fileName
				}
			}
		}
	}

	if avatarPath != "" && user.Avatar.Valid && user.Avatar.String != avatarPath {
		oldFileObj := filepath.Join("public", strings.TrimPrefix(user.Avatar.String, "/"))

		if _, err := os.Stat(oldFileObj); err == nil {
			err = os.Remove(oldFileObj)
			if err != nil {
				log.Println("WARNING: Gagal menghapus avatar lama fisik:", err)
			}
		}
	}

	err := models.UpdateUserProfile(uc.env.DB, userID, address, phone, avatarPath)
	if err != nil {
		uc.RenderError(w, r, http.StatusInternalServerError, "Terjadi Kesalahan Pada Sistem, Silahkan Hubungi Administrator.")
		log.Printf("CRITICAL ERROR path=%s, err=%v", r.URL.Path, err)
		return
	}

	changes := map[string]interface{}{
        "address": address,
        "phone":   phone,
    }

    if avatarPath != "" {
        changes["avatar_url"] = uc.env.BaseURL + "/avatar/" + strconv.Itoa(userID)
    }

    services.SendUserUpdateWebhook(uc.env, user, changes)

	session, _ := uc.env.Store.Get(r, uc.env.SessionName)
    session.AddFlash("Data berhasil diperbarui!")
    session.Save(r, w)

	http.Redirect(w, r, "/profile/edit", http.StatusFound)
}

// GET /avatar/{id} untuk konsumsi luar
func (uc *UserController) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetIDStr := vars["userID"]
	targetID, _ := strconv.Atoi(targetIDStr)

	user, err := models.FindUserByID(uc.env.DB, targetID)
	if err != nil {
		http.ServeFile(w, r, "./public/uploads/avatars/default.png")
		return
	}

	if !user.Avatar.Valid || user.Avatar.String == "" {

		if user.GoogleAvatar.Valid && user.GoogleAvatar.String != "" {

			resp, err := http.Get(user.GoogleAvatar.String)
			if err == nil && resp.StatusCode == 200 {
				defer resp.Body.Close()
				w.Header().Set("Content-Type", "image/jpeg")
				io.Copy(w, resp.Body)
				return
			}
		}

		http.ServeFile(w, r, "./public/uploads/avatars/default.png")
		return

	}

	if user.Avatar.String == user.GoogleAvatar.String {

		resp, err := http.Get(user.GoogleAvatar.String)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			w.Header().Set("Content-Type", "image/jpeg")
			io.Copy(w, resp.Body)
			return
		}
	}

	uploadDir := "./public/uploads/avatars"
	pattern := filepath.Join(uploadDir, fmt.Sprintf("user-%d-avatar-*.jpg", targetID))

	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		http.ServeFile(w, r, "./public/uploads/avatars/default.png")
		return
	}

	filePath := matches[0]

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.Header().Set("Cache-Control", "max-age=60, must-revalidate")

	http.ServeFile(w, r, filePath)
}

func (ac *UserController) RenderError(w http.ResponseWriter, r *http.Request, code int, message string) {
    w.WriteHeader(code)
    
    data := map[string]interface{}{
        "Code":    code,
        "Message": message,
    }
    
    ac.views.RenderPage(w, r, "error", data)
}

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
	"sso-portal-v3/config"
	"sso-portal-v3/models"
	"sso-portal-v3/views"
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
	}

	uc.views.RenderPage(w, r, "edit-profile", data)
}

// HandleProfileUpdate memproses data form edit profil
func (uc *UserController) HandleProfileUpdate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("UserLogin").(*models.FullUser)
	userID := user.ID

	// 1. Parse form (multipart karena ada file upload)
	// Batasi ukuran upload
	r.ParseMultipartForm(10 << 20) // 10 MB

	// 2. Ambil data teks
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
			// Decode base64 string
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
	// Hanya jalankan ini jika avatarPath masih kosong (artinya Skenario A tidak jalan/error)
	if avatarPath == "" {
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()

			ext := filepath.Ext(header.Filename)
			// Tambahkan timestamp agar browser tidak cache gambar lama
			fileName := fmt.Sprintf("user-%d-avatar-%d%s", userID, time.Now().UnixNano(), ext)
			savePath := filepath.Join(saveDir, fileName)

			dst, err := os.Create(savePath)
			if err == nil {
				defer dst.Close()
				if _, err := io.Copy(dst, file); err == nil {
					avatarPath = "/uploads/avatars/" + fileName
					log.Printf("INFO: User %d update avatar via FILE UPLOAD", userID)
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
			} else {
				log.Println("SUCCESS: Avatar lama dihapus:", oldFileObj)
			}
		}
	}

	// 4. Update database
	err := models.UpdateUserProfile(uc.env.DB, userID, address, phone, avatarPath)
	if err != nil {
		log.Printf("ERROR: Gagal update profil user %d: %v", userID, err)
		http.Error(w, "Gagal mengupdate profil", http.StatusInternalServerError)
		return
	}

	// 5. Redirect kembali ke dashboard dengan pesan sukses
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

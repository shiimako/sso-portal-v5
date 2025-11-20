// file: controllers/usercontroller/usercontroller.go
package usercontroller

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sso-portal-v3/handlers"
	"sso-portal-v3/models"
)

type UserController struct {
	env *handlers.Env
}

func NewUserController(env *handlers.Env) *UserController {
	return &UserController{env: env}
}

// ShowProfileForm menampilkan halaman edit profil pengguna
func (uc *UserController) ShowProfileForm(w http.ResponseWriter, r *http.Request) {
	session, _ := uc.env.Store.Get(r, uc.env.SessionName)
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user, err := models.FindUserByID(uc.env.DB, fmt.Sprintf("%d", userID))
	if err != nil {
		http.Error(w, "Gagal mengambil data pengguna", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User": user,
	}
	
	// Asumsi kamu pakai sistem template yang sama
	uc.env.Templates.ExecuteTemplate(w, "edit_profile.html", data) 
}

// HandleProfileUpdate memproses data form edit profil
func (uc *UserController) HandleProfileUpdate(w http.ResponseWriter, r *http.Request) {
	session, _ := uc.env.Store.Get(r, uc.env.SessionName)
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// 1. Parse form (multipart karena ada file upload)
	// Batasi ukuran upload
	r.ParseMultipartForm(10 << 20) // 10 MB

	// 2. Ambil data teks
	name := r.FormValue("name")
	address := r.FormValue("address")
	phone := r.FormValue("phone")
	var avatarPath string = ""

	// 3. Handle File Upload (Avatar)
	file, header, err := r.FormFile("avatar")
	if err == nil { // Jika ada file baru yang diupload
		defer file.Close()
		
		// Buat nama file unik, misal: user_8_avatar.png
		ext := filepath.Ext(header.Filename)
		fileName := fmt.Sprintf("user_%d_avatar%s", userID, ext)
		
		// Tentukan path simpan (pastikan folder 'static/uploads/avatars' ada)
		// Sesuaikan path 'static' ini dengan cara kamu menyajikan file statis
		savePath := filepath.Join("static", "uploads", "avatars", fileName)

		// Buat file tujuan
		dst, err := os.Create(savePath)
		if err != nil {
			http.Error(w, "Gagal membuat file di server", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Salin file yang diupload ke file tujuan
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
			return
		}
		
		// Simpan path relatif untuk database
		avatarPath = "/" + savePath // Path harus bisa diakses via URL
		log.Printf("INFO: User %d mengupload avatar baru ke: %s", userID, avatarPath)

	} else if err != http.ErrMissingFile {
		// Jika ada error selain "file tidak ada", tampilkan error
		log.Printf("ERROR: Gagal membaca form file: %v", err)
	}

	// 4. Update database
	err = models.UpdateUserProfile(uc.env.DB, userID, name, address, phone, avatarPath)
	if err != nil {
		log.Printf("ERROR: Gagal update profil user %d: %v", userID, err)
		http.Error(w, "Gagal mengupdate profil", http.StatusInternalServerError)
		return
	}

	// 5. Redirect kembali ke dashboard dengan pesan sukses
	http.Redirect(w, r, "/dashboard?success=Profil berhasil diupdate", http.StatusFound)
}
package main

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"sso-portal-v3/config"
	"sso-portal-v3/controllers/admincontroller"
	"sso-portal-v3/controllers/authcontroller"
	"sso-portal-v3/controllers/dashboardcontroller"
	"sso-portal-v3/controllers/redirectcontroller"
	"sso-portal-v3/handlers"
	"sso-portal-v3/middleware"
	"sso-portal-v3/views"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables dari file .env
	err := godotenv.Load()
    if err != nil {
        log.Println("Peringatan: Gagal memuat file .env")
    }

	// Register tipe data kustom untuk session
	gob.Register([]map[string]interface{}{}) // Untuk menyimpan slice of map di session

	// Inisialisasi database
	db, err := config.InitDB()
    if err != nil {
        log.Fatalf("Gagal inisialisasi database: %v", err)
    }
    defer db.Close()

	// Inisialisasi session store
	sessionStore := config.InitSessionStore()

	// Inisialisasi template renderer
	templates, err := views.InitTemplates()
	if err != nil {
		log.Fatalf("Gagal inisialisasi template: %v", err)
	}

	// Inisialisasi environment untuk handler
	env := &handlers.Env{
		DB:           	db,
		Store: 			sessionStore,
		Templates:    	templates,
		SessionName:    os.Getenv("SESSION_NAME"),
		BaseURL: 	  	os.Getenv("APP_BASE_URL"),
	}

	// Inisialisasi Google OAuth2 config
	oauthConfig := config.InitGoogleOAuthConfig(env.BaseURL)
	env.GoogleOAuthConfig = oauthConfig

	// Inisialisasi controller
	authCtrl := authcontroller.NewAuthController(env)
	dashboardCtrl := dashboardcontroller.NewDashboardController(env)
	adminCtrl := admincontroller.NewAdminController(env)
	redirectCtrl := redirectcontroller.NewRedirectController(env)

	// Inisialisasi router
	r := mux.NewRouter()

	// Route untuk halaman login
	r.HandleFunc("/", authCtrl.ShowLoginPage).Methods("GET")
	// Route untuk login dengan Google
	r.HandleFunc("/auth/google/login", authCtrl.LoginWithGoogle).Methods("GET")
	// Route untuk callback dari Google
	r.HandleFunc("/auth/google/callback", authCtrl.GoogleCallback).Methods("GET")
	// Route untuk select role
	r.HandleFunc("/select-role", authCtrl.SelectRolePage).Methods("GET") 
	// Route untuk set role
	r.HandleFunc("/set-role", authCtrl.SelectRoleHandler).Methods("GET")
	// Route untuk logout
	r.HandleFunc("/logout", authCtrl.Logout).Methods("GET")

	// Route untuk dashboard
	r.HandleFunc("/dashboard", dashboardCtrl.Index).Methods("GET")

	// Rute Admin
	// Buat subrouter khusus untuk semua path yang diawali '/admin'
	adminRouter := r.PathPrefix("/admin").Subrouter()
	// Terapkan middleware "satpam" ke semua rute di dalam subrouter ini
	adminRouter.Use(middleware.AdminMiddleware(env))

	// Buat handler admin dashboard sementara untuk testing
	adminRouter.HandleFunc("/dashboard", adminCtrl.Dashboard).Methods("GET")

	// ===================================
	// REDIRECT MANAGEMENT
	// ====================================
	r.HandleFunc("/redirect", redirectCtrl.RedirectToApp).Methods("GET")

	// ===================================
	// USER MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/users", adminCtrl.ListUsers).Methods("GET")
	adminRouter.HandleFunc("/users/detail/{id}", adminCtrl.DetailUser).Methods("GET")
	adminRouter.HandleFunc("/users/new", adminCtrl.NewUserForm).Methods("GET")
	adminRouter.HandleFunc("/users/create", adminCtrl.CreateUser).Methods("POST")
	adminRouter.HandleFunc("/users/edit/{id}", adminCtrl.EditUserForm).Methods("GET")
	adminRouter.HandleFunc("/users/update/{id}", adminCtrl.UpdateUser).Methods("POST")
	adminRouter.HandleFunc("/users/delete/{id}", adminCtrl.DeleteUser).Methods("POST")

	// ===================================
	// APPLICATION MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/applications", adminCtrl.ListApplications).Methods("GET")
	adminRouter.HandleFunc("/applications/detail/{id}", adminCtrl.DetailApplication).Methods("GET")
	adminRouter.HandleFunc("/applications/new", adminCtrl.NewApplicationForm).Methods("GET")
	adminRouter.HandleFunc("/applications/create", adminCtrl.CreateApplication).Methods("POST")
	adminRouter.HandleFunc("/applications/edit/{id}", adminCtrl.EditApplicationForm).Methods("GET")
	adminRouter.HandleFunc("/applications/update/{id}", adminCtrl.UpdateApplication).Methods("POST")
	adminRouter.HandleFunc("/applications/delete/{id}", adminCtrl.DeleteApplication).Methods("POST")

	// ===================================
	// ROLE MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/roles", adminCtrl.ListRoles).Methods("GET")
	// adminRouter.HandleFunc("/roles/new", adminCtrl.NewRoleForm).Methods("GET")
	// adminRouter.HandleFunc("/roles/create", adminCtrl.CreateRole).Methods("POST")
	// adminRouter.HandleFunc("/roles/edit/{id}", adminCtrl.EditRoleForm).Methods("GET")
	// adminRouter.HandleFunc("/roles/update/{id}", adminCtrl.UpdateRole).Methods("POST")
	// adminRouter.HandleFunc("/roles/delete/{id}", adminCtrl.DeleteRole).Methods("POST")

	port := os.Getenv("PORT")
	log.Printf("Server berjalan di http://localhost:%s", port)
	err = http .ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
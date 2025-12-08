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
	"sso-portal-v3/controllers/usercontroller"
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
	env := &config.Env{
		DB:          db,
		Store:       sessionStore,
		Templates:   templates,
		SessionName: os.Getenv("SESSION_NAME"),
		BaseURL:     os.Getenv("APP_BASE_URL"),
	}

	viewEngine := views.NewViews(env)

	// Inisialisasi Google OAuth2 config
	oauthConfig := config.InitGoogleOAuthConfig(env.BaseURL)
	env.GoogleOAuthConfig = oauthConfig

	// Load RSA keys untuk JWT
	if err := config.LoadKeys(); err != nil {
    log.Fatalf("Gagal memuat RSA keys: %v", err)
	}

	// Inisialisasi controller
	authCtrl := authcontroller.NewAuthController(env)
	dashboardCtrl := dashboardcontroller.NewDashboardController(env, viewEngine)
	adminCtrl := admincontroller.NewAdminController(env, viewEngine)
	redirectCtrl := redirectcontroller.NewRedirectController(env, viewEngine)
	userCtrl := usercontroller.NewUserController(env, viewEngine)

	// Inisialisasi router
	r := mux.NewRouter()

	uploadsFs := http.FileServer(http.Dir("./public/uploads"))
	r.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", uploadsFs),
	)
	staticFs := http.FileServer(http.Dir("./public/static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFs))

	// ===================================
	// AUTHENTICATION ROUTES
	// ====================================
	r.HandleFunc("/", authCtrl.ShowLoginPage).Methods("GET")
	r.HandleFunc("/auth/google/login", authCtrl.LoginWithGoogle).Methods("GET")
	r.HandleFunc("/auth/google/callback", authCtrl.GoogleCallback).Methods("GET")
	r.HandleFunc("/logout", authCtrl.Logout).Methods("GET")

	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.GlobalAuthMiddleware(env))
	// ===================================
	// DASHBOARD ROUTES
	// ====================================
	protected.HandleFunc("/dashboard", dashboardCtrl.Index).Methods("GET")

	// ===================================
	// USER PROFILE ROUTES
	// ====================================
	protected.HandleFunc("/profile/edit", userCtrl.ShowProfileForm).Methods("GET")
	protected.HandleFunc("/profile/update", userCtrl.HandleProfileUpdate).Methods("POST")
	r.HandleFunc("/avatar/{userID}", userCtrl.ServeAvatar).Methods("GET")

	// ===================================
	// REDIRECT MANAGEMENT
	// ====================================
	protected.HandleFunc("/redirect", redirectCtrl.RedirectToApp).Methods("GET")

	// ===================================
	// ADMIN ROUTES
	// ====================================
	adminRouter := protected.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminMiddleware(env))
	adminRouter.HandleFunc("/dashboard", adminCtrl.Dashboard).Methods("GET")

	// ===================================
	// USER MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/users", adminCtrl.ListUsers).Methods("GET")
	adminRouter.HandleFunc("/users/detail/{id}", adminCtrl.DetailUser).Methods("GET")

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
	// LOG MANAGEMENT
	// ====================================
	adminRouter.HandleFunc("/sync/stream", adminCtrl.StreamUserSync).Methods("GET")
	adminRouter.HandleFunc("/sync-logs", adminCtrl.SyncLogsPage).Methods("GET")

	adminRouter.HandleFunc("/jurusan", adminCtrl.ListJurusan).Methods("GET")
	adminRouter.HandleFunc("/jurusan/sync/stream", adminCtrl.StreamJurusanSync).Methods("GET")

	adminRouter.HandleFunc("/prodi", adminCtrl.ListProdi).Methods("GET")
	adminRouter.HandleFunc("/prodi/sync/stream", adminCtrl.StreamProdiSync).Methods("GET")

	adminRouter.HandleFunc("/jabatan", adminCtrl.ListJabatan).Methods("GET")
	adminRouter.HandleFunc("/jabatan/sync/stream", adminCtrl.StreamJabatanSync).Methods("GET")

	adminRouter.HandleFunc("/roles", adminCtrl.ListRoles).Methods("GET")
	adminRouter.HandleFunc("/roles/sync/stream", adminCtrl.StreamRoleSync).Methods("GET")

	port := os.Getenv("PORT")
	log.Printf("Server berjalan di http://localhost:%s", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

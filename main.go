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
		DB:          db,
		Store:       sessionStore,
		Templates:   templates,
		SessionName: os.Getenv("SESSION_NAME"),
		BaseURL:     os.Getenv("APP_BASE_URL"),
	}

	// Inisialisasi Google OAuth2 config
	oauthConfig := config.InitGoogleOAuthConfig(env.BaseURL)
	env.GoogleOAuthConfig = oauthConfig

	// Inisialisasi controller
	authCtrl := authcontroller.NewAuthController(env)
	dashboardCtrl := dashboardcontroller.NewDashboardController(env)
	adminCtrl := admincontroller.NewAdminController(env)
	redirectCtrl := redirectcontroller.NewRedirectController(env)
	userCtrl := usercontroller.NewUserController(env)

	// Inisialisasi router
	r := mux.NewRouter()

	
	// ===================================
	// AUTHENTICATION ROUTES
	// ====================================
	r.HandleFunc("/", authCtrl.ShowLoginPage).Methods("GET")
	r.HandleFunc("/auth/google/login", authCtrl.LoginWithGoogle).Methods("GET")
	r.HandleFunc("/auth/google/callback", authCtrl.GoogleCallback).Methods("GET")
	r.HandleFunc("/select-role", authCtrl.SelectRolePage).Methods("GET")
	r.HandleFunc("/set-role", authCtrl.SelectRoleHandler).Methods("GET")
	r.HandleFunc("/logout", authCtrl.Logout).Methods("GET")

	// ===================================
	// DASHBOARD ROUTES
	// ====================================
	r.HandleFunc("/dashboard", dashboardCtrl.Index).Methods("GET")

	// ===================================
	// USER PROFILE ROUTES
	// ====================================
	r.HandleFunc("/profile/edit", userCtrl.ShowProfileForm).Methods("GET")
	r.HandleFunc("/profile/update", userCtrl.HandleProfileUpdate).Methods("POST")

	// ===================================
	// REDIRECT MANAGEMENT
	// ====================================
	r.HandleFunc("/redirect", redirectCtrl.RedirectToApp).Methods("GET")


	// ===================================
	// ADMIN ROUTES
	// ====================================
	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AdminMiddleware(env))
	adminRouter.HandleFunc("/dashboard", adminCtrl.Dashboard).Methods("GET")

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
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
